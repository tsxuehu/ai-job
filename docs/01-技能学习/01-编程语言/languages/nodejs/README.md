# Node.js / TypeScript：从语言语义到事件驱动后端

Node.js 项目同时涉及三层：JavaScript 的运行时语义、TypeScript 的编译期类型系统、Node.js 的事件循环与标准库。学习时必须分清哪一层提供了什么保证。

贯穿案例是订单服务：TypeScript 建模订单和支付合同，Promise 表达异步结果，Stream 处理数据，AbortSignal 传播取消，worker_threads 隔离 CPU 工作，测试和 profile 提供证据。

> 本机示例基线为 Node.js 24。生产项目应固定 Node LTS 与 TypeScript 版本。先运行[订单示例](../../examples/README.md#可运行订单示例)。

> 学完使用 [Node.js 分类面试题](../../面试题/nodejs/README.md)检查表达。

## 1. 最终要建立的心智模型

- 原始值、对象引用、浅复制和闭包分别共享什么？
- TypeScript 类型为什么在运行时消失，外部 JSON 由谁验证？
- `type`、`interface`、class、判别联合和泛型各适合什么模型？
- Promise、microtask、timer 和 IO callback 按什么阶段运行？
- `await` 在哪里暂停，为什么不会自动取消底层 IO？
- 单个进程何时可以处理大量 IO，何时会被 CPU/同步调用阻塞？
- Stream 背压如何阻止内存被高速生产者打满？
- Buffer、string、编码和二进制协议如何转换？
- ESM、CommonJS、package exports 和构建产物如何保持一致？
- V8 heap、native memory、handle、listener 和 Promise 如何形成泄漏？

## 2. 贯穿案例的增长路线

```text
TypeScript 建模 Order
→ interface/判别联合表达合同与状态
→ Promise 连接 PaymentGateway
→ Map/Array/AsyncIterable 处理订单流
→ Error + cause 保留失败链
→ Stream/HTTP/数据库跨越 IO 边界
→ AbortSignal + 并发限制管理任务
→ node:test/typecheck/profile 提供证据
→ workspace/package exports 形成可交付工程
```

## 3. 十五章学习顺序

| 顺序 | 章节 | 读完必须会做 |
| ---: | --- | --- |
| 1 | [基础语法](01-基础语法.md) | 区分 JS、TS、Node、ESM 和进程入口 |
| 2 | [变量、数据与类型](02-变量数据与类型.md) | 解释值、引用、浅复制、null 和运行时数据 |
| 3 | [表达式与逻辑控制](03-表达式与逻辑控制.md) | 使用严格相等、收窄、判别联合和循环 |
| 4 | [函数与作用域](04-函数与作用域.md) | 设计函数类型、闭包、this、泛型和 async 返回值 |
| 5 | [面向对象与抽象](05-面向对象与抽象.md) | 选择 type/interface/class/组合/泛型 |
| 6 | [容器与迭代](06-容器与迭代.md) | 使用 Map/Set/Iterable/AsyncIterable 和背压 |
| 7 | [错误处理](07-错误处理.md) | 管理同步异常、Promise rejection、cause 和边界 |
| 8 | [内存与资源](08-内存与资源.md) | 分析 V8 heap、handle、listener 和资源关闭 |
| 9 | [IO 与网络](09-IO与网络.md) | 管理 Buffer、Stream、HTTP、JSON、数据库和超时 |
| 10 | [并发与异步](10-并发与异步.md) | 解释事件循环，管理并发、取消和 CPU 工作 |
| 11 | [运行时与性能](11-运行时与性能.md) | 解释运行时对象、Reflect/Proxy、V8、GC 和事件循环 |
| 12 | [测试与工程实践](12-测试与工程实践.md) | 使用 node:test、类型检查、真实边界和 profile |
| 13 | [工程化](13-工程化.md) | 组织 workspace、package、exports、依赖、CI 和发布 |
| 14 | [项目注意事项](14-项目注意事项.md) | 审查未等待 Promise、阻塞、取消、模块和安全风险 |
| 15 | [编译、运行与调试命令](15-编译运行与调试命令.md) | 使用 npm scripts、typecheck、Inspector、profile 和产物验证 |

## 4. 从其他语言迁移时要修正的直觉

| 旧直觉 | Node.js / TypeScript 中要改成 |
| --- | --- |
| Java/C++ 类型在运行时存在 | 多数 TS 类型被擦除，边界必须运行时校验 |
| Go goroutine/Java thread | Promise 不是线程；普通 JS 默认在事件循环线程运行 |
| Python async task | 语法接近但取消协议、Stream 和库行为不同 |
| Java checked exception/Go error | TS 默认不在函数类型中表达抛出的错误 |
| 多线程服务按请求占线程 | Node 擅长非阻塞 IO，但 CPU 和同步 API 会阻塞所有请求 |

## 5. 七天高密度路线

| 天 | 阅读与编码 | 当天证据 |
| ---: | --- | --- |
| 1 | 01—03：JS 值、TS 类型、模块和收窄 | 画对象图并验证类型擦除后的运行结果 |
| 2 | 04—06：函数、抽象、集合与异步迭代 | 完成订单模型、支付接口和订单流 |
| 3 | 07—09：错误、资源、Stream、HTTP、数据库 | 超时能取消底层操作，Stream 遵守背压 |
| 4 | 10：事件循环、Promise、并发限制、worker | 压测阻塞和非阻塞两种实现 |
| 5 | 11：V8、GC、event-loop delay 和 profile | 用证据解释一次延迟或内存异常 |
| 6 | 12：node:test、集成测试和性能工具 | 固定质量命令可重复执行 |
| 7 | 13—15：workspace、命令、构建和生产规则 | 案例可构建、调试、启动、测试和优雅停机 |

## 6. 每章的学习动作

先运行 JS/TS 代码，再分别问：TypeScript 编译期检查了什么、运行时剩下什么、事件循环何时执行、资源如何结束。用测试、event-loop delay、heap snapshot 和 CPU profile 验证，不根据 `async` 关键字猜测并发行为。

## 7. 综合实战与证据

实现有界订单 API + Worker：外部 JSON 运行时校验，支付/存储可替换，Stream 有背压，任务有并发上限和 AbortSignal，CPU 工作不会阻塞主事件循环；提交 node:test、真实 adapter 测试、heap/CPU profile、event-loop delay 数据与可复现锁文件。

## 8. 权威资料怎么配合

- [Node.js Learn](https://nodejs.org/en/learn/getting-started/introduction-to-nodejs)：建立事件循环、异步 IO、诊断和性能路径；
- [Node.js API](https://nodejs.org/api/)：查询 Stream、Buffer、HTTP、AbortSignal、worker_threads 等运行时契约；
- [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html)：按 narrowing、function、object、generics、module 理解类型系统；
- [JavaScript Language Reference](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference)：查证 ECMAScript 值、表达式、对象和 Promise 语义。

TypeScript Handbook 解释编译期，Node 文档解释运行时，JavaScript reference 解释语言本身；三者必须分开使用。
