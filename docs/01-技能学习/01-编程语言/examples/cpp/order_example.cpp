#include <cassert>
#include <cstdint>
#include <stdexcept>
#include <string>
#include <utility>

enum class OrderStatus { pending, paid };

struct PaymentReceipt {
    std::string id;
};

class Order {
public:
    Order(std::string id, std::int64_t amount_cents)
        : id_{std::move(id)}, amount_cents_{amount_cents} {
        if (id_.empty()) {
            throw std::invalid_argument{"order id must not be empty"};
        }
        if (amount_cents_ <= 0) {
            throw std::invalid_argument{"amount must be positive"};
        }
    }

    const std::string& id() const { return id_; }
    std::int64_t amount_cents() const { return amount_cents_; }
    OrderStatus status() const { return status_; }
    const std::string& payment_id() const { return payment_id_; }

    void mark_paid(std::string payment_id) {
        if (status_ != OrderStatus::pending) {
            throw std::logic_error{"order already paid"};
        }
        if (payment_id.empty()) {
            throw std::invalid_argument{"payment id must not be empty"};
        }
        payment_id_ = std::move(payment_id);
        status_ = OrderStatus::paid;
    }

private:
    std::string id_;
    std::int64_t amount_cents_;
    OrderStatus status_{OrderStatus::pending};
    std::string payment_id_;
};

class PaymentGateway {
public:
    virtual ~PaymentGateway() = default;
    virtual PaymentReceipt pay(
        const std::string& order_id,
        std::int64_t amount_cents
    ) = 0;
};

class FakeGateway final : public PaymentGateway {
public:
    PaymentReceipt pay(const std::string&, std::int64_t) override {
        ++calls;
        return PaymentReceipt{"pay-1"};
    }
    int calls{};
};

void checkout(Order& order, PaymentGateway& gateway) {
    auto receipt = gateway.pay(order.id(), order.amount_cents());
    order.mark_paid(std::move(receipt.id));
}

int main() {
    bool rejected = false;
    try {
        Order invalid{"o-0", 0};
    } catch (const std::invalid_argument&) {
        rejected = true;
    }
    assert(rejected);

    Order order{"o-1", 100};
    FakeGateway gateway;
    checkout(order, gateway);
    assert(gateway.calls == 1);
    assert(order.status() == OrderStatus::paid);
    assert(order.payment_id() == "pay-1");
}
