# Python：从对象模型到生产服务

Python 语法短，但动态语义并不简单。这套资料面向有后端经验的开发者，重点建立“名称绑定对象”的模型、数据模型协议、可变性、异常、迭代器、asyncio、解释器和包工程。

贯穿案例仍是订单服务：dataclass 表达订单，Protocol 抽象支付，异常表达失败，文件/HTTP/数据库形成边界，asyncio 处理并发，pytest/profile 形成证据。

> 本机示例基线为 Python 3.12。先运行[订单示例](../../examples/README.md#可运行订单示例)，新版本特性应由项目版本约束决定。

> 学完使用 [Python 分类面试题](../../面试题/python/README.md)检查表达。

## 1. 最终要建立的心智模型

- 变量为什么不是盒子，而是名称到对象的绑定？
- 赋值、浅复制、深复制和函数传参分别产生什么对象图？
- `is`、`==`、`__eq__`、hash 和可变性如何影响集合？
- class、dataclass、Protocol、ABC 和泛型分别解决什么问题？
- iterable、iterator、generator 如何通过协议协作？
- 异常链、context manager 和 `ExceptionGroup` 如何表达失败与清理？
- 类型标注为什么不自动验证 HTTP/JSON 等运行时输入？
- threading、multiprocessing、asyncio 应按什么负载选择？
- GIL 具体限制什么，又不保证什么？
- import、虚拟环境、distribution package 和应用目录分别是什么？

## 2. 贯穿案例的增长路线

```text
名称绑定 Order 对象
→ dataclass 和方法维护状态
→ Protocol 隔离 PaymentGateway
→ list/dict/generator 处理订单流
→ exception + context manager 保证失败清理
→ HTTP/JSON/数据库进入运行时边界
→ asyncio TaskGroup 管理并发任务
→ pytest/type checker/profile 提供证据
→ pyproject + src layout 形成可交付工程
```

## 3. 十四章学习顺序

| 顺序 | 章节 | 读完必须会做 |
| ---: | --- | --- |
| 1 | [基础语法](01-基础语法.md) | 运行模块并理解缩进、import 和入口 |
| 2 | [变量、数据与类型](02-变量数据与类型.md) | 画出名称、对象、共享、复制和可变性 |
| 3 | [表达式与逻辑控制](03-表达式与逻辑控制.md) | 正确使用真值、推导式、match 和循环 |
| 4 | [函数与作用域](04-函数与作用域.md) | 设计参数、闭包、装饰器和 LEGB 作用域 |
| 5 | [面向对象与抽象](05-面向对象与抽象.md) | 使用数据模型、Protocol、ABC 和泛型建模 |
| 6 | [容器与迭代](06-容器与迭代.md) | 理解容器复杂度、iterator 和 generator |
| 7 | [错误处理](07-错误处理.md) | 设计异常层次、链、边界和清理动作 |
| 8 | [内存与资源](08-内存与资源.md) | 解释引用计数/GC，并用上下文管理资源 |
| 9 | [IO 与网络](09-IO与网络.md) | 管理编码、流、HTTP、JSON 和数据库边界 |
| 10 | [并发与异步](10-并发与异步.md) | 选择线程/进程/asyncio 并管理任务生命周期 |
| 11 | [运行时与性能](11-运行时与性能.md) | 解释解释器、字节码、GIL、分配和热点 |
| 12 | [测试与工程实践](12-测试与工程实践.md) | 使用 pytest、类型检查、profile 和基准 |
| 13 | [工程化](13-工程化.md) | 组织 package、pyproject、依赖、入口和发布 |
| 14 | [项目注意事项](14-项目注意事项.md) | 审查可变默认值、阻塞 async、import 和安全风险 |

## 4. 从其他语言迁移时要修正的直觉

| 旧直觉 | Python 中要改成 |
| --- | --- |
| C++/Go 值变量 | Python 名称通常绑定对象，赋值不复制对象 |
| Java/TypeScript 声明类型 | 标注默认不做运行时强制，边界仍需验证 |
| Java interface | Protocol 可以按结构匹配；ABC 提供运行时名义关系 |
| Node async/await | 语法相似，但 asyncio 的任务、取消和库生态不同 |
| Go goroutine/Java thread | async task 只在事件循环协作；CPU 工作不能因此并行 |

## 5. 七天高密度路线

| 天 | 阅读与编码 | 当天证据 |
| ---: | --- | --- |
| 1 | 01—03：对象、绑定、可变性、控制流 | 画对象图并解释浅复制结果 |
| 2 | 04—06：函数、数据模型、Protocol、迭代器 | 完成订单模型和惰性订单流 |
| 3 | 07—09：异常、资源、编码、HTTP、数据库 | 错误保留 cause，资源能稳定关闭 |
| 4 | 10：线程、进程、asyncio、取消 | TaskGroup 任务可失败、取消和等待 |
| 5 | 11：解释器、GIL、内存和 profile | 用 profile 解释一个热点或保留路径 |
| 6 | 12：pytest、类型检查、属性测试和 benchmark | 固定质量命令可重复执行 |
| 7 | 13—14：pyproject、入口、依赖和生产规则 | 案例可安装、测试、启动和部署 |

## 6. 每章的学习动作

先在 REPL/测试中运行；画名称与对象关系；修改可变/不可变输入并预测；用类型检查、测试、tracemalloc/profile 验证；最后把动态便利对应的运行时责任写清楚。

## 7. 综合实战与证据

实现有界订单 API + Worker：运行时校验外部数据，Protocol 隔离依赖，HTTP/数据库资源由 context manager 管理，async 任务可取消和等待；提交 pytest、真实边界测试、类型检查、内存/CPU profile 与可复现 `pyproject.toml`。

## 8. 权威资料怎么配合

- [The Python Tutorial](https://docs.python.org/3/tutorial/)：从解释器、控制流、数据结构进入模块、IO、异常和类；
- [Python Data Model](https://docs.python.org/3/reference/datamodel.html)：理解对象、属性、调用、迭代和特殊方法协议；
- [Fluent Python, 2nd Edition](https://www.oreilly.com/library/view/fluent-python-2nd/9781492056348/)：按数据结构、函数即对象、类协议、控制流和元编程建立 Pythonic 设计；
- [Python Packaging User Guide](https://packaging.python.org/)：处理 pyproject、构建、distribution 和发布。

Tutorial 建立路线，Data Model 解释原理，Fluent Python 训练惯用抽象，Packaging Guide 负责交付。
