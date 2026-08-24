import assert from "node:assert/strict";
import test from "node:test";

type Status = "pending" | "paid";
type PaymentReceipt = Readonly<{ id: string }>;

class Order {
  #status: Status = "pending";
  #paymentId?: string;
  readonly id: string;
  readonly amountCents: number;

  constructor(id: string, amountCents: number) {
    if (id.length === 0) throw new TypeError("order id must not be empty");
    if (amountCents <= 0) throw new RangeError("amount must be positive");
    this.id = id;
    this.amountCents = amountCents;
  }

  get status(): Status {
    return this.#status;
  }

  get paymentId(): string | undefined {
    return this.#paymentId;
  }

  markPaid(paymentId: string): void {
    if (this.#status !== "pending") throw new Error("order already paid");
    if (paymentId.length === 0) throw new TypeError("payment id must not be empty");
    this.#paymentId = paymentId;
    this.#status = "paid";
  }
}

interface PaymentGateway {
  pay(orderId: string, amountCents: number): Promise<PaymentReceipt>;
}

async function checkout(order: Order, gateway: PaymentGateway): Promise<void> {
  const receipt = await gateway.pay(order.id, order.amountCents);
  order.markPaid(receipt.id);
}

test("rejects invalid amount and supports a fake gateway", async () => {
  assert.throws(() => new Order("o-0", 0), RangeError);

  let calls = 0;
  const fake: PaymentGateway = {
    async pay(): Promise<PaymentReceipt> {
      calls += 1;
      return { id: "pay-1" };
    },
  };
  const order = new Order("o-1", 100);
  await checkout(order, fake);
  assert.equal(calls, 1);
  assert.equal(order.status, "paid");
  assert.equal(order.paymentId, "pay-1");
});
