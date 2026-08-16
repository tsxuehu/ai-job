# Rust 面试

## 高频考点

- 所有权、移动、借用和生命周期。
- `String`/`&str`，`Box`/`Rc`/`Arc`。
- `Option`、`Result`、trait、泛型与 trait object。
- `Send`、`Sync`、`Arc<Mutex<T>>` 和线程安全。
- async/await、Future、Pin 和运行时。
- unsafe、FFI 和安全抽象边界。

## 必须能回答

所有权怎样避免内存错误和数据竞争？为什么阻塞任务会拖慢 async runtime，应如何隔离？
