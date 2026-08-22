public final class OrderExample {
    enum Status { PENDING, PAID }

    static final class Order {
        private final String id;
        private final long amountCents;
        private Status status = Status.PENDING;

        Order(String id, long amountCents) {
            if (amountCents <= 0) throw new IllegalArgumentException("amount must be positive");
            this.id = id;
            this.amountCents = amountCents;
        }

        void markPaid() {
            if (status != Status.PENDING) throw new IllegalStateException("order already paid");
            status = Status.PAID;
        }
    }

    interface PaymentGateway {
        void pay(String orderId, long amountCents);
    }

    static void checkout(Order order, PaymentGateway gateway) {
        gateway.pay(order.id, order.amountCents);
        order.markPaid();
    }

    public static void main(String[] args) {
        boolean rejected = false;
        try {
            new Order("o-0", 0);
        } catch (IllegalArgumentException expected) {
            rejected = true;
        }
        assert rejected;

        var calls = new int[] {0};
        PaymentGateway fake = (id, amount) -> calls[0]++;
        var order = new Order("o-1", 100);
        checkout(order, fake);
        assert calls[0] == 1;
        assert order.status == Status.PAID;
    }
}

