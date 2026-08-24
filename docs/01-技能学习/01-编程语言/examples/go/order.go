package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidID     = errors.New("order ID must not be empty")
	ErrInvalidAmount = errors.New("amount must be positive")
	ErrAlreadyPaid   = errors.New("order already paid")
)

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
)

type Order struct {
	id          string
	amountCents int64
	status      Status
	paymentID   string
}

func NewOrder(id string, amountCents int64) (*Order, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidID
	}
	if amountCents <= 0 {
		return nil, ErrInvalidAmount
	}
	return &Order{id: id, amountCents: amountCents, status: StatusPending}, nil
}

func (o *Order) MarkPaid(paymentID string) error {
	if o.status != StatusPending {
		return ErrAlreadyPaid
	}
	o.status = StatusPaid
	o.paymentID = paymentID
	return nil
}

func (o Order) ID() string         { return o.id }
func (o Order) AmountCents() int64 { return o.amountCents }
func (o Order) Status() Status     { return o.status }
func (o Order) PaymentID() string  { return o.paymentID }

type PaymentReceipt struct {
	ID string
}

type PaymentGateway interface {
	Pay(ctx context.Context, orderID string, amountCents int64) (PaymentReceipt, error)
}

func Checkout(ctx context.Context, order *Order, gateway PaymentGateway) error {
	receipt, err := gateway.Pay(ctx, order.ID(), order.AmountCents())
	if err != nil {
		return fmt.Errorf("pay order %s: %w", order.ID(), err)
	}
	if err := order.MarkPaid(receipt.ID); err != nil {
		return fmt.Errorf("mark order %s paid: %w", order.ID(), err)
	}
	return nil
}
