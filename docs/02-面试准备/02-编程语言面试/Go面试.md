# Go 面试

## 高频考点

- slice 扩容、map、interface、泛型和值/指针语义。
- goroutine 调度、channel、select、mutex 和内存模型。
- channel 关闭责任及关闭后的读写行为。
- context 的超时、取消和传递规则。
- 逃逸分析、GC、defer 和 error 处理。
- race detector、benchmark、pprof 和优雅停机。

## 必须能回答

goroutine 泄漏有哪些典型路径？如何用指标和 pprof 证明泄漏来源，并保证流式请求取消后资源被释放？
