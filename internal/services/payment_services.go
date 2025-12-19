package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	mt "GameStoreAPI/external/midtrans"
	"GameStoreAPI/internal/repository"

	"github.com/google/uuid"
	midtrans "github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type PaymentService struct {
	PaymentRepo       *repository.PaymentRepository
	OrderRepo         *repository.OrderRepository
	CustomerGamesRepo *repository.CustomerGamesRepository
	CartRepo          *repository.CartRepository
	Snap              *snap.Client
}

func NewPaymentService(
	pr *repository.PaymentRepository,
	or *repository.OrderRepository,
	cgr *repository.CustomerGamesRepository,
	cr *repository.CartRepository,
	snap *snap.Client,
) *PaymentService {
	return &PaymentService{
		PaymentRepo:       pr,
		OrderRepo:         or,
		CustomerGamesRepo: cgr,
		CartRepo:          cr,
		Snap:              snap,
	}
}

func (s *PaymentService) CreateSnapPayment(
	ctx context.Context,
	orderID int64,
	authID int64,
) (string, error) {

	// 🔹 Resolve authID → customerID
	customerID, err := s.CartRepo.GetCustomerID(ctx, authID)
	if err != nil {
		return "", errors.New("customer not found")
	}

	order, err := s.OrderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return "", errors.New("order not found")
	}

	// 🔒 Ownership check
	if order.CustomerID != customerID {
		return "", errors.New("forbidden")
	}

	if order.TotalPrice == nil {
		return "", errors.New("order not ready for payment")
	}

	if order.OrderStatus != "PendingPayment" {
		return "", errors.New("order cannot be paid")
	}

	existing, err := s.PaymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return "", err
	}
	if existing != nil && existing.PaymentStatus == "Pending" {
		return "", errors.New("payment already exists")
	}

	externalRef := fmt.Sprintf("ORDER-%d-%s", orderID, uuid.NewString())

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  externalRef,
			GrossAmt: int64(*order.TotalPrice),
		},
	}

	resp, snapErr := s.Snap.CreateTransaction(req)
	if snapErr != nil {
		return "", snapErr
	}

	payload, _ := json.Marshal(resp)

	_, err = s.PaymentRepo.CreatePending(
		ctx,
		orderID,
		int64(*order.TotalPrice),
		"midtrans",
		externalRef,
		payload,
	)
	if err != nil {
		return "", err
	}

	return resp.RedirectURL, nil
}

func (s *PaymentService) HandleMidtransNotification(ctx context.Context, payload map[string]interface{}) error {

	orderIDStr, ok := payload["order_id"].(string)
	if !ok {
		return errors.New("missing order_id")
	}

	// Extract internal order ID from ORDER-{id}-UUID
	var orderID int64
	if _, err := fmt.Sscanf(orderIDStr, "ORDER-%d-", &orderID); err != nil {
		return errors.New("invalid order reference")
	}

	order, err := s.OrderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.OrderStatus == "Paid" {
		// already processed → safely ignore
		return nil
	}

	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signature, _ := payload["signature_key"].(string)

	if !mt.VerifySignature(
		orderIDStr,
		statusCode,
		grossAmount,
		signature,
		os.Getenv("MIDTRANS_SERVER_KEY"),
	) {
		return errors.New("invalid signature")
	}

	transactionStatus, _ := payload["transaction_status"].(string)
	fraudStatus, _ := payload["fraud_status"].(string)

	switch transactionStatus {

	case "settlement":
		return s.finalizePayment(ctx, orderID, payload)

	case "capture":
		if fraudStatus == "accept" {
			return s.finalizePayment(ctx, orderID, payload)
		}

	case "expire", "cancel", "deny":
		return s.markPaymentFailed(ctx, orderID, payload)
	}

	return nil
}

func (s *PaymentService) HandleMidtransWebhook(
	ctx context.Context,
	payload map[string]interface{},
) error {

	orderIDStr, _ := payload["order_id"].(string)
	status, _ := payload["transaction_status"].(string)

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		return errors.New("invalid order id")
	}

	if status != "settlement" && status != "capture" {
		return nil // ignore pending, cancel, expire
	}

	// 1️⃣ mark payment + order paid
	if err := s.markPaymentPaid(ctx, orderID, payload); err != nil {
		return err
	}

	// 2️⃣ grant owned games
	return s.grantGamesFromOrder(ctx, orderID)
}

func (s *PaymentService) markPaymentPaid(
	ctx context.Context,
	orderID int64,
	payload map[string]interface{},
) error {

	data, _ := json.Marshal(payload)

	transactionID, _ := payload["transaction_id"].(string)
	paymentType, _ := payload["payment_type"].(string)

	tx, err := s.PaymentRepo.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.PaymentRepo.MarkPaidTx(
		ctx,
		tx,
		orderID,
		transactionID,
		paymentType,
		data,
	); err != nil {
		return err
	}

	if err := s.OrderRepo.MarkPaidTx(ctx, tx, orderID); err != nil {
		return err
	}

	// grant games here

	return tx.Commit(ctx)
}

func (s *PaymentService) markPaymentFailed(
	ctx context.Context,
	orderID int64,
	payload map[string]interface{},
) error {

	// Idempotency guard
	order, err := s.OrderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order.OrderStatus == "Paid" || order.OrderStatus == "Failed" {
		return nil
	}

	data, _ := json.Marshal(payload)

	tx, err := s.PaymentRepo.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Mark order failed
	if err := s.OrderRepo.MarkFailed(ctx, orderID); err != nil {
		return err
	}

	// Mark payment failed (optional but good)
	if err := s.PaymentRepo.MarkFailed(ctx, orderID, data); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PaymentService) finalizePayment(
	ctx context.Context,
	orderID int64,
	payload map[string]interface{},
) error {

	// 1️⃣ Idempotency guard
	order, err := s.OrderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.OrderStatus == "Paid" {
		return nil
	}

	// 2️⃣ Extract Midtrans fields
	provider := "midtrans"

	providerRef, _ := payload["transaction_id"].(string)

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 3️⃣ Get purchased games
	gameIDs, err := s.OrderRepo.GetGameIDsByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	// 4️⃣ Transaction
	tx, err := s.PaymentRepo.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 5️⃣ Mark payment as paid (repo expects full data)
	if err := s.PaymentRepo.MarkPaidTx(
		ctx,
		tx,
		orderID,
		provider,
		providerRef,
		rawPayload,
	); err != nil {
		return err
	}

	// 6️⃣ Grant ownership
	if err := s.CustomerGamesRepo.CreateCustomerGamesTx(
		ctx,
		tx,
		order.CustomerID,
		gameIDs,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PaymentService) grantGamesFromOrder(
	ctx context.Context,
	orderID int64,
) error {

	order, err := s.OrderRepo.GetOrderByID(ctx, orderID)
	if err != nil {
		return err
	}

	gameIDs, err := s.OrderRepo.GetGameIDsByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	return s.CustomerGamesRepo.InsertPurchased(
		ctx,
		order.CustomerID,
		gameIDs,
	)
}
