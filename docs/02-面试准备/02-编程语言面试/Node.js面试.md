# Node.js / TypeScript 面试

## 高频考点

- 事件循环阶段，Promise、microtask、`process.nextTick` 和 timer 的顺序。
- `async/await` 错误传播、取消和未处理 Promise。
- CPU 密集任务为何阻塞，worker/process 如何选择。
- Stream、Buffer 和背压；SSE 断连后的资源释放。
- TypeScript 类型擦除，为什么仍需运行时 schema 校验。
- 模块系统、依赖、框架中间件和请求生命周期。
- 事件监听器、定时器、闭包和堆内存泄漏的定位。

## 必须能回答

如何测量事件循环延迟？如果模型流式接口导致连接和内存持续增长，你如何确认影响范围、定位并修复？
