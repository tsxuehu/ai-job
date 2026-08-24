# 可运行订单示例

这些示例只使用语言标准能力，验证四件事：订单编号和金额合法、支付实现可以替换、支付成功返回回执、订单拿到支付编号后才改变状态。

| 语言 | 运行命令 |
|---|---|
| C++ | `c++ -std=c++20 cpp/order_example.cpp -o /tmp/order_cpp && /tmp/order_cpp` |
| Go | `cd go && go test ./...` |
| Java | `javac -d /tmp java/OrderExample.java && java -ea -cp /tmp OrderExample` |
| Node.js/TypeScript | `node --test nodejs/order-example.test.ts` |
| Python | `python3 -m unittest python/test_order_example.py` |

示例保持单文件或最小 module，便于直接运行。真实服务的目录、依赖、数据库、HTTP 和部署方式参见各语言的第 13 章“工程化”。
