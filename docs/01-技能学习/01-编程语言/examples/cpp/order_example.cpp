#include <cassert>
#include <cstdint>
#include <stdexcept>
#include <string>
#include <utility>

enum class OrderStatus { pending, paid };

class Order {
public:
    Order(std::string id, std::int64_t amount_cents)
        : id_{std::move(id)}, amount_cents_{amount_cents} {
        if (amount_cents_ <= 0) {
            throw std::invalid_argument{"amount must be positive"};
        }
    }

    const std::string& id() const { return id_; }
    std::int64_t amount_cents() const { return amount_cents_; }
    OrderStatus status() const { return status_; }

    void mark_paid() {
        if (status_ != OrderStatus::pending) {
            throw std::logic_error{"order already paid"};
        }
        status_ = OrderStatus::paid;
    }

private:
    std::string id_;
    std::int64_t amount_cents_;
    OrderStatus status_{OrderStatus::pending};
};

class PaymentGateway {
public:
    virtual ~PaymentGateway() = default;
    virtual void pay(const std::string& order_id, std::int64_t amount_cents) = 0;
};

class FakeGateway final : public PaymentGateway {
public:
    void pay(const std::string&, std::int64_t) override { ++calls; }
    int calls{};
};

void checkout(Order& order, PaymentGateway& gateway) {
    gateway.pay(order.id(), order.amount_cents());
    order.mark_paid();
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
}

