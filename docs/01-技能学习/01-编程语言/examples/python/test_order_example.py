import unittest
from dataclasses import dataclass, field
from enum import Enum
from typing import Protocol


class Status(Enum):
    PENDING = "pending"
    PAID = "paid"


@dataclass
class Order:
    id: str
    amount_cents: int
    status: Status = field(default=Status.PENDING, init=False)

    def __post_init__(self) -> None:
        if self.amount_cents <= 0:
            raise ValueError("amount must be positive")

    def mark_paid(self) -> None:
        if self.status is not Status.PENDING:
            raise RuntimeError("order already paid")
        self.status = Status.PAID


class PaymentGateway(Protocol):
    def pay(self, order_id: str, amount_cents: int) -> None: ...


def checkout(order: Order, gateway: PaymentGateway) -> None:
    gateway.pay(order.id, order.amount_cents)
    order.mark_paid()


class FakeGateway:
    def __init__(self) -> None:
        self.calls = 0

    def pay(self, order_id: str, amount_cents: int) -> None:
        self.calls += 1


class OrderTest(unittest.TestCase):
    def test_checkout(self) -> None:
        with self.assertRaises(ValueError):
            Order("o-0", 0)

        order = Order("o-1", 100)
        gateway = FakeGateway()
        checkout(order, gateway)
        self.assertEqual(gateway.calls, 1)
        self.assertIs(order.status, Status.PAID)


if __name__ == "__main__":
    unittest.main()

