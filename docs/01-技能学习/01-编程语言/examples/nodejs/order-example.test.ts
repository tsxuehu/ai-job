import assert from "node:assert/strict";
import test from "node:test";

type Status = "pending" | "paid";

class Order {
  #status: Status = "pending";
  readonly id: string;
  readonly amountCents: number;

  constructor(id: string, amountCents: number) {
    if (amountCents <= 0) throw new RangeError("amount must be positive");
    this.id = id;
    this.amountCents = amountCents;
  }

  get status(): Status {
    return this.#status;
  }

  markPaid(): void {
    if (this.#status !== "pending") throw new Error("order already paid");
    this.#status = "paid";
  }
}

interface PaymentGateway {
  pay(orderId: string, amountCents: number): Promise<void>;
}

async function checkout(order: Order, gateway: PaymentGateway): Promise<void> {
  await gateway.pay(order.id, order.amountCents);
  order.markPaid();
}

test("rejects invalid amount and supports a fake gateway", async () => {
  assert.throws(() => new Order("o-0", 0), RangeError);

  let calls = 0;
  const fake: PaymentGateway = {
    async pay(): Promise<void> {
      calls += 1;
    },
  };
  const order = new Order("o-1", 100);
  await checkout(order, fake);
  assert.equal(calls, 1);
  assert.equal(order.status, "paid");
});
