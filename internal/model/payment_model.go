package model

import "time"

type Payment struct {
	PaymentID       int64      `db:"paymentid" json:"payment_id"`
	OrderID         int64      `db:"orderid" json:"order_id"`
	PaymentMethodID int64      `db:"paymentmethodid" json:"payment_method_id"`
	AmountPaid      float64    `db:"amountpaid" json:"amount_paid"`
	PaymentStatus   string     `db:"paymentstatus" json:"payment_status"`
	PaymentProvider *string    `db:"paymentprovider" json:"payment_provider"`
	ProviderRef     *string    `db:"providerref" json:"provider_ref"`
	ProviderPayload []byte     `db:"providerpayload" json:"provider_payload"`
	CreatedAt       time.Time  `db:"createdat" json:"created_at"`
	PaidAt          *time.Time `db:"paidat" json:"paid_at"`
}
