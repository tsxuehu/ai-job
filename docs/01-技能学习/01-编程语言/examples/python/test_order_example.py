import unittest
from dataclasses import dataclass, field
from enum import Enum
from typing import Optional, Protocol


class Status(Enum):
    PENDING = "pending"
    PAID = "paid"


@dataclass(frozen=True)
class PaymentReceipt:
    id: str


@dataclass
class Order:
    id: str
    amount_cents: int
    status: Status = field(default=Status.PENDING, init=False)
    payment_id: Optional[str] = field(default=None, init=False)

    def __post_init__(self) -> None:
        if not self.id:
            raise ValueError("order id must not be empty")
        if self.amount_cents <= 0:
            raise ValueError("amount must be positive")

    def mark_paid(self, payment_id: str) -> None:
        if self.status is not Status.PENDING:
            raise RuntimeError("order already paid")
        if not payment_id:
            raise ValueError("payment id must not be empty")
        self.payment_id = payment_id
        self.status = Status.PAID


class PaymentGateway(Protocol):
    def pay(self, order_id: str, amount_cents: int) -> PaymentReceipt: ...


def checkout(order: Order, gateway: PaymentGateway) -> None:
    receipt = gateway.pay(order.id, order.amount_cents)
    order.mark_paid(receipt.id)


class FakeGateway:
    def __init__(self) -> None:
        self.calls = 0

    def pay(self, order_id: str, amount_cents: int) -> PaymentReceipt:
        self.calls += 1
        return PaymentReceipt("pay-1")


class OrderTest(unittest.TestCase):
    def test_checkout(self) -> None:
        with self.assertRaises(ValueError):
            Order("o-0", 0)

        order = Order("o-1", 100)
        gateway = FakeGateway()
        checkout(order, gateway)
        self.assertEqual(gateway.calls, 1)
        self.assertIs(order.status, Status.PAID)
        self.assertEqual(order.payment_id, "pay-1")


if __name__ == "__main__":
    unittest.main()
