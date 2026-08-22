package order

import "errors"

var (
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
}

func NewOrder(id string, amountCents int64) (*Order, error) {
	if amountCents <= 0 {
		return nil, ErrInvalidAmount
	}
	return &Order{id: id, amountCents: amountCents, status: StatusPending}, nil
}

func (o *Order) MarkPaid() error {
	if o.status != StatusPending {
		return ErrAlreadyPaid
	}
	o.status = StatusPaid
	return nil
}

type PaymentGateway interface {
	Pay(orderID string, amountCents int64) error
}

func Checkout(order *Order, gateway PaymentGateway) error {
	if err := gateway.Pay(order.id, order.amountCents); err != nil {
		return err
	}
	return order.MarkPaid()
}

