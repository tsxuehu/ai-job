package order

import (
	"errors"
	"testing"
)

type fakeGateway struct{ calls int }

func (f *fakeGateway) Pay(string, int64) error {
	f.calls++
	return nil
}

func TestOrderCheckout(t *testing.T) {
	if _, err := NewOrder("o-0", 0); !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("got %v, want ErrInvalidAmount", err)
	}

	order, err := NewOrder("o-1", 100)
	if err != nil {
		t.Fatal(err)
	}
	gateway := &fakeGateway{}
	if err := Checkout(order, gateway); err != nil {
		t.Fatal(err)
	}
	if gateway.calls != 1 || order.status != StatusPaid {
		t.Fatalf("calls=%d status=%s", gateway.calls, order.status)
	}
}

