package repository

import (
	"context"
	"errors"

	"GameStoreAPI/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	DB *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{DB: db}
}

func (r *PaymentRepository) CreatePending(
	ctx context.Context,
	orderID int64,
	amount int64,
	provider string,
	providerRef string,
	payload []byte,
) (int64, error) {

	var paymentID int64
	q := `
		INSERT INTO payments
			(orderid, amountpaid, paymentstatus, paymentprovider, providerref, providerpayload, createdat)
		VALUES
			($1, $2, 'Pending', $3, $4, $5, NOW())
		RETURNING paymentid
	`
	err := r.DB.QueryRow(
		ctx, q,
		orderID, amount, provider, providerRef, payload,
	).Scan(&paymentID)

	return paymentID, err
}

func (r *PaymentRepository) GetByOrderID(
	ctx context.Context,
	orderID int64,
) (*model.Payment, error) {

	var p model.Payment

	q := `
		SELECT paymentid, orderid, amountpaid, paymentstatus,
		       paymentprovider, providerref, providerpayload,
		       createdat, paidat
		FROM payments
		WHERE orderid=$1
	`

	err := r.DB.QueryRow(ctx, q, orderID).Scan(
		&p.PaymentID,
		&p.OrderID,
		&p.AmountPaid,
		&p.PaymentStatus,
		&p.PaymentProvider,
		&p.ProviderRef,
		&p.ProviderPayload,
		&p.CreatedAt,
		&p.PaidAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

/* NON TRANSACTION
func (r *PaymentRepository) MarkPaid(
	ctx context.Context,
	orderID int64,
	payload []byte,
) error {

	q := `
		UPDATE payments
		SET paymentstatus='Paid',
		    providerpayload=$2,
		    paidat=NOW()
		WHERE orderid=$1
		  AND paymentstatus='Pending'
	`
	_, err := r.DB.Exec(ctx, q, orderID, payload)
	return err
}
*/

func (r *PaymentRepository) MarkPaidTx(
	ctx context.Context,
	tx pgx.Tx,
	orderID int64,
	providerRef string,
	provider string,
	payload []byte,
) error {

	_, err := tx.Exec(ctx, `
		UPDATE payments
		SET paymentstatus='Paid',
		    providerref=$2,
		    paymentprovider=$3,
		    providerpayload=$4,
		    paidat=NOW()
		WHERE orderid=$1 AND paymentstatus='Pending'
	`, orderID, providerRef, provider, payload)

	return err
}

func (r *PaymentRepository) MarkFailed(
	ctx context.Context,
	orderID int64,
	payload []byte,
) error {
	_, err := r.DB.Exec(ctx, `
		UPDATE payments
		SET paymentstatus='Failed',
		    providerpayload=$2
		WHERE orderid=$1
		  AND paymentstatus='Pending'
	`, orderID, payload)
	return err
}
