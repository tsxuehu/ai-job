public final class OrderExample {
    enum Status { PENDING, PAID }

    record PaymentReceipt(String id) {}

    static final class Order {
        private final String id;
        private final long amountCents;
        private Status status = Status.PENDING;
        private String paymentId;

        Order(String id, long amountCents) {
            if (id == null || id.isBlank()) throw new IllegalArgumentException("order id must not be blank");
            if (amountCents <= 0) throw new IllegalArgumentException("amount must be positive");
            this.id = id;
            this.amountCents = amountCents;
        }

        void markPaid(String paymentId) {
            if (status != Status.PENDING) throw new IllegalStateException("order already paid");
            if (paymentId == null || paymentId.isBlank()) {
                throw new IllegalArgumentException("payment id must not be blank");
            }
            this.paymentId = paymentId;
            status = Status.PAID;
        }
    }

    interface PaymentGateway {
        PaymentReceipt pay(String orderId, long amountCents);
    }

    static void checkout(Order order, PaymentGateway gateway) {
        var receipt = gateway.pay(order.id, order.amountCents);
        order.markPaid(receipt.id());
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
        PaymentGateway fake = (id, amount) -> {
            calls[0]++;
            return new PaymentReceipt("pay-1");
        };
        var order = new Order("o-1", 100);
        checkout(order, fake);
        assert calls[0] == 1;
        assert order.status == Status.PAID;
        assert order.paymentId.equals("pay-1");
    }
}
