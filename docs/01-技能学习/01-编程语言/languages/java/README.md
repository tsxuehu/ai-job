# Java：从类型系统到 JVM 工程

这套资料面向已有多语言经验的后端开发者。目标是建立 Java 的值/引用模型、类与泛型、异常、Java Memory Model、JVM 执行和企业工程边界。

贯穿案例是订单服务：Order 维护状态，PaymentGateway 隔离支付，HTTP/JDBC 跨越边界，Executor 或虚拟线程执行任务，JUnit/JFR 提供证据，构建工具交付应用。

> 本机示例基线为 Java 17；虚拟线程等新特性会明确标注需要 Java 21+。先运行[订单示例](../../examples/README.md#可运行订单示例)。

> 学完使用 [Java 分类面试题](../../面试题/java/README.md)检查表达，不把 Spring 注解记忆当作 Java 能力。

## 1. 最终要建立的心智模型

- 基本类型变量和引用变量分别保存什么？方法参数到底复制了什么？
- `==`、`equals`、`hashCode` 如何共同决定集合行为？
- class、record、enum、sealed interface 分别适合什么模型？
- 接口调用如何分派，泛型擦除会留下哪些边界？
- checked/unchecked exception、异常链和资源关闭如何设计？
- 对象可达为什么不等于业务资源已经关闭？
- JMM 中 happens-before、`volatile`、锁和线程安全分别解决什么？
- 字节码如何经过类加载、解释/JIT、GC 变成运行行为？
- Spring 代理、事务和 ORM 为什么会改变普通方法调用的直觉？
- package、JPMS module、Maven/Gradle module 和 Git 仓库分别是什么？

## 2. 贯穿案例的增长路线

```text
创建合法 Order
→ 方法和封装维护状态
→ PaymentGateway interface 实现多态
→ 泛型集合保存订单
→ exception 表达失败并保留 cause
→ HTTP/JDBC 进入真实边界
→ Executor/CompletableFuture 管理任务
→ JUnit/race-style stress/JFR 提供证据
→ feature package + 构建模块形成工程
```

## 3. 十四章学习顺序

| 顺序 | 章节 | 读完必须会做 |
| ---: | --- | --- |
| 1 | [基础语法](01-基础语法.md) | 从源码解释编译、字节码、类和入口 |
| 2 | [变量、数据与类型](02-变量数据与类型.md) | 解释基本值、引用、对象、String 和相等性 |
| 3 | [表达式与逻辑控制](03-表达式与逻辑控制.md) | 正确使用转换、switch、模式和循环 |
| 4 | [方法与作用域](04-函数与作用域.md) | 设计参数、返回值、重载、lambda 和作用域 |
| 5 | [面向对象与抽象](05-面向对象与抽象.md) | 使用 class/record/interface/sealed/泛型建模 |
| 6 | [容器与迭代](06-容器与迭代.md) | 选择集合并维护 equals/hashCode/并发边界 |
| 7 | [错误处理](07-错误处理.md) | 设计异常类别、cause、边界转换和清理 |
| 8 | [内存与资源](08-内存与资源.md) | 区分 GC 可达性和 AutoCloseable 生命周期 |
| 9 | [IO 与网络](09-IO与网络.md) | 管理字符/字节、HTTP、JSON、JDBC 和超时 |
| 10 | [并发与异步](10-并发与异步.md) | 解释 JMM，管理线程池、Future 和取消 |
| 11 | [运行时与性能](11-运行时与性能.md) | 解释类加载、JIT、GC、分配和 profile |
| 12 | [测试与工程实践](12-测试与工程实践.md) | 使用 JUnit、集成测试、JFR/JMH 和诊断工具 |
| 13 | [工程化](13-工程化.md) | 组织 feature、构建模块、依赖、CI 和发布 |
| 14 | [项目注意事项](14-项目注意事项.md) | 审查代理、事务、ORM、线程池和类加载风险 |

## 4. 从其他语言迁移时要修正的直觉

| 旧直觉 | Java 中要改成 |
| --- | --- |
| C++ 值对象/析构 | 普通对象变量保存引用；GC 不提供确定析构时刻 |
| Python/TypeScript 结构化协议 | Java interface 是名义类型，实现关系显式声明 |
| Go 小接口和显式 error | Java 常用接口、异常与框架代理，但异常边界仍要设计 |
| Node 单事件循环 | Java 线程可并行；共享状态必须遵守 JMM |
| TypeScript 泛型 | Java 泛型多数通过擦除实现，运行时不能直接获得 `T` |

## 5. 七天高密度路线

| 天 | 阅读与编码 | 当天证据 |
| ---: | --- | --- |
| 1 | 01—03：类型、引用、相等性、控制流 | 画对象图并解释每次参数传递 |
| 2 | 04—06：方法、OO、泛型、集合 | 完成订单模型和两个支付实现 |
| 3 | 07—09：异常、资源、HTTP、JDBC | 失败路径保留 cause，资源均可证明关闭 |
| 4 | 10：JMM、Executor、Future、取消 | 实现有界 Worker 并完成停机测试 |
| 5 | 11：类加载、JIT、GC、JFR | 用 JFR/profile 解释一个真实现象 |
| 6 | 12：单元、集成、并发和 JMH | 固定命令可重复产生测试与性能结果 |
| 7 | 13—14：feature、Spring、事务、交付 | 整理为可发布服务并做故障演练 |

## 6. 每章的学习动作

先运行案例，画引用和对象关系；再修改代码预测编译/运行结果；用测试、线程 dump、JFR 或 JMH 验证；最后写出与 C++/Go/Python/Node 的语义差异和生产后果。

## 7. 综合实战与证据

实现订单 API + Worker：领域对象保持不变量，支付和存储可替换，HTTP/JDBC 资源可关闭，任务有界并可取消，异常能映射稳定协议；同时提交单元/真实数据库/停机测试、JFR 记录、一次 JMH 或服务压测以及构建依赖图。

## 8. 权威资料怎么配合

- [dev.java Learn](https://dev.java/learn/)：按语言基础、集合、IO、并发和 JVM 专题学习；
- [Java Language Specification](https://docs.oracle.com/javase/specs/jls/se25/html/index.html)：裁决类型、表达式、类和并发语义；
- [Effective Java](https://www.informit.com/store/effective-java-9780134686097)：用条目、理由和后果训练 API、泛型、异常、并发和序列化设计；
- [Java Virtual Machine Specification](https://docs.oracle.com/javase/specs/jvms/se25/html/index.html)：深入字节码、类文件和运行时结构。

教程建立顺序，Effective Java 建立工程判断，JLS/JVMS 用于精确查证。
