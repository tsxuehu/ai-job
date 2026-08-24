package order

import (
	"context"
	"errors"
	"testing"
)

type fakeGateway struct {
	calls   int
	receipt PaymentReceipt
	err     error
}

func (f *fakeGateway) Pay(context.Context, string, int64) (PaymentReceipt, error) {
	f.calls++
	return f.receipt, f.err
}

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		amountCents int64
		wantErr     error
	}{
		{name: "valid", id: "order-1", amountCents: 100},
		{name: "empty ID", id: " ", amountCents: 100, wantErr: ErrInvalidID},
		{name: "zero amount", id: "order-1", amountCents: 0, wantErr: ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOrder(tt.id, tt.amountCents)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewOrder() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckoutMarksOrderPaid(t *testing.T) {
	order, err := NewOrder("order-1", 100)
	if err != nil {
		t.Fatal(err)
	}
	gateway := &fakeGateway{receipt: PaymentReceipt{ID: "payment-1"}}

	if err := Checkout(context.Background(), order, gateway); err != nil {
		t.Fatal(err)
	}
	if gateway.calls != 1 || order.Status() != StatusPaid || order.PaymentID() != "payment-1" {
		t.Fatalf(
			"calls=%d status=%s paymentID=%q",
			gateway.calls,
			order.Status(),
			order.PaymentID(),
		)
	}
}

func TestCheckoutPreservesGatewayError(t *testing.T) {
	order, err := NewOrder("order-1", 100)
	if err != nil {
		t.Fatal(err)
	}
	gatewayErr := errors.New("gateway unavailable")
	gateway := &fakeGateway{err: gatewayErr}

	err = Checkout(context.Background(), order, gateway)
	if !errors.Is(err, gatewayErr) {
		t.Fatalf("Checkout() error = %v, want wrapped gateway error", err)
	}
	if order.Status() != StatusPending {
		t.Fatalf("status = %q, want %q", order.Status(), StatusPending)
	}
}
