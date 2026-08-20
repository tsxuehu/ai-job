# Go

Go 的核心是用较小的语言和统一工具链构建可读、可并发、易部署的服务。快速学习时重点理解值复制、slice/map 的共享语义、interface、显式错误、goroutine 生命周期和 `context` 取消传播。

> 开始真实项目之前，先阅读 [Go 项目特性与注意事项](项目注意事项.md)；建立完整仓库时使用 [Go 企业级项目结构](企业级项目结构.md)。

## 快速路线

| 顺序 | 分类 | 掌握结果 |
| ---: | --- | --- |
| 1 | [基础语法](01-基础语法.md) | 能理解 package、声明、导出和工具链约束 |
| 2 | [变量、数据与类型](02-变量数据与类型.md) | 能理解值复制、零值、slice/map 和字符串语义 |
| 3 | [表达式与逻辑控制](03-表达式与逻辑控制.md) | 能正确使用 if、switch、for、range 和多值表达式 |
| 4 | [函数与作用域](04-函数与作用域.md) | 能处理多返回值、闭包、遮蔽、defer 与共享 |
| 5 | [自定义类型与抽象](05-自定义类型与抽象.md) | 能用 struct、method、interface、组合和泛型抽象 |
| 6 | [容器与迭代](06-容器与迭代.md) | 能管理 slice/map 的容量、共享、迭代和并发边界 |
| 7 | [错误处理](07-错误处理.md) | 能设计错误链、协议映射和 panic 恢复边界 |
| 8 | [内存与资源](08-内存与资源.md) | 能理解 GC、逃逸、分配和业务资源关闭 |
| 9 | [模块化、依赖与 IO](09-模块化依赖与IO.md) | 能按能力划分 package 并治理依赖和 IO 边界 |
| 10 | [并发与异步](10-并发与异步.md) | 能管理 goroutine、channel、context、锁和背压 |
| 11 | [运行时与性能](11-运行时与性能.md) | 能理解调度、GC 并用 pprof/trace 定位热点 |
| 12 | [测试与工程实践](12-测试与工程实践.md) | 能使用 test、race、fuzz、CI 与生产观测 |

## 七天快速上手

1. 第 1 天：完成变量、类型、slice、map、字符串和控制流程。
2. 第 2 天：使用函数、struct、method、interface 和泛型建模。
3. 第 3 天：训练显式错误、defer、资源关闭和 nil 边界。
4. 第 4 天：建立 Go module，完成文件、JSON 和 HTTP 服务。
5. 第 5 天：实现 goroutine、channel、context 和有界并发。
6. 第 6 天：补齐单元、集成、并发、超时和关闭测试。
7. 第 7 天：运行 race detector、benchmark、pprof 和 trace。

七天后应能独立写出小型服务，但仍需通过故障注入和生产指标训练调度、GC 与并发判断。

## 需要学习

- slice、map、interface、泛型、值/指针语义、error 和 defer。
- goroutine、channel、select、mutex、context。
- 调度器、GC、逃逸分析和内存模型。
- HTTP/RPC、中间件、连接池、超时和优雅停机。
- 单元测试、benchmark、race detector 和 pprof。
- 并发限制、取消传播、流式响应和 goroutine 泄漏治理。

## 验收

开发一个高并发网关或任务服务，并使用 race detector、benchmark 和 pprof 定位竞争、阻塞与性能问题。

建议验收项目：实现一个有界任务 API，包含参数校验、存储接口、超时、取消、并发限制、优雅停机、指标和测试；能够证明没有明显 goroutine 泄漏和 data race。
