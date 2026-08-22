# 五门语言的内存与资源：订单查询结束后谁负责释放

> 目标：看懂对象内存可以自动管理，但数据库连接、文件、锁和后台任务仍必须由代码明确关闭。

统一案例：导出订单报表。

```text
1. 从连接池取得数据库连接
2. 查询订单
3. 写入报表文件
4. 中途任何一步失败，都要释放连接并关闭文件
```

最重要的问题只有两个：

1. 谁拥有这个资源？
2. 在哪条执行路径上释放？

---

## 1. 一张表先看结论

| 语言 | 普通对象内存 | 外部资源的主要管理方式 |
|---|---|---|
| C++ | 栈对象、手动/智能指针，无通用 GC | RAII、析构函数、智能指针 |
| Go | GC | `defer Close()`，显式取消和等待 |
| Java | JVM GC | try-with-resources、`AutoCloseable` |
| Node.js/TS | V8 GC | `try/finally`、Stream/句柄的显式关闭 |
| Python | 引用计数 + 循环 GC | `with` / `async with` 上下文管理器 |

GC 只解决“对象内存何时回收”的一部分问题。它不会可靠地替你提交事务、归还连接、释放锁或停止后台任务。

---

## 2. C++：资源绑定到对象生命周期

```cpp
void export_orders(Database& database, const std::filesystem::path& path) {
    auto connection = database.acquire(); // RAII 连接对象
    auto rows = connection.query("SELECT id, total FROM orders");

    std::ofstream output{path};
    if (!output) {
        throw ReportError{"cannot open report file"};
    }

    for (const auto& row : rows) {
        output << row.id << "," << row.total << "\n";
    }
}
```

函数正常返回或抛出异常时，局部对象按逆序析构：

```text
output 析构并关闭文件
rows 析构并释放查询结果
connection 析构并归还连接
```

这就是 RAII：创建对象时取得资源，对象析构时释放资源。

堆对象优先使用：

```cpp
std::unique_ptr<T> // 单一所有者
std::shared_ptr<T> // 确实需要共享所有权时
std::weak_ptr<T>   // 观察 shared_ptr，避免所有权环
```

不要用裸 `new/delete` 表达普通所有权。裸指针和引用更适合表示“不拥有，只借用”，但必须保证被借用对象活得足够久。

---

## 3. Go：GC 管内存，`defer` 管关闭

```go
func ExportOrders(
    ctx context.Context,
    db *sql.DB,
    path string,
) (err error) {
    rows, err := db.QueryContext(ctx, "SELECT id, total FROM orders")
    if err != nil {
        return err
    }
    defer rows.Close()

    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer func() {
        if closeErr := file.Close(); err == nil {
            err = closeErr
        }
    }()

    for rows.Next() {
        // Scan 后写入文件。
    }
    return rows.Err()
}
```

`defer` 在当前函数返回前执行，适合紧跟资源创建写关闭逻辑。

需要注意：

- `sql.DB` 通常是长期连接池，不要每个请求创建和关闭；
- `sql.Rows` 属于一次查询，必须关闭；
- 写文件时 `Close` 也可能失败，不能永远忽略；
- 在大循环中反复 `defer` 会推迟到整个函数结束，应抽成小函数或显式关闭；
- goroutine 也有生命周期，启动方必须能取消并等待它结束。

Go 的逃逸分析决定对象放栈还是堆。业务代码通常先写清楚所有权和生命周期，再通过 profile 判断分配问题。

---

## 4. Java：try-with-resources 保证关闭

```java
void exportOrders(DataSource dataSource, Path path) throws SQLException, IOException {
    try (
        Connection connection = dataSource.getConnection();
        PreparedStatement statement = connection.prepareStatement(
            "SELECT id, total FROM orders"
        );
        ResultSet rows = statement.executeQuery();
        BufferedWriter output = Files.newBufferedWriter(path)
    ) {
        while (rows.next()) {
            output.write(rows.getString("id"));
            output.write(",");
            output.write(rows.getString("total"));
            output.newLine();
        }
    }
}
```

实现 `AutoCloseable` 的资源离开 `try` 时会按逆序关闭，即使中途抛异常也一样。

需要注意：

- GC 回收 Java 对象，不等于及时归还数据库连接；
- `finalize` 不是资源管理方案；
- 连接池、Executor 和 HTTP client 通常属于应用级资源；
- ResultSet、Stream、临时文件句柄通常属于一次操作；
- 关闭时出现的 suppressed exception 仍应在诊断中保留。

---

## 5. Node.js / TypeScript：用 `finally` 保证归还

```ts
async function exportOrders(
  pool: Pool,
  path: string,
  signal: AbortSignal
): Promise<void> {
  const client = await pool.connect();
  const output = createWriteStream(path);

  try {
    const rows = await client.query({
      text: "SELECT id, total FROM orders",
      signal,
    });

    for (const row of rows) {
      if (!output.write(`${row.id},${row.total}\n`)) {
        await once(output, "drain");
      }
    }
  } finally {
    output.end();
    client.release();
  }
}
```

JavaScript 对象由 V8 回收，但以下资源会让进程继续存活或耗尽容量：

- 数据库连接；
- socket 和 HTTP keep-alive 连接；
- Stream、文件句柄；
- timer、listener、Worker；
- 未完成且无法取消的异步操作。

Stream 关闭和错误处理比示例更复杂，生产代码应等待 `finish/close` 并处理 `error`。重点是：取得资源的代码必须有明确的 `finally` 或生命周期管理器。

Buffer 还可能使用 V8 堆外内存，只看 JavaScript heap 不一定能解释进程 RSS。

---

## 6. Python：使用 `with` 和 `async with`

同步文件使用上下文管理器：

```python
def write_report(path: Path, rows: Iterable[OrderRow]) -> None:
    with path.open("w", encoding="utf-8") as output:
        for row in rows:
            output.write(f"{row.id},{row.total}\n")
```

异步数据库连接使用异步上下文管理器：

```python
async def export_orders(pool: Pool, path: Path) -> None:
    async with pool.acquire() as connection:
        rows = await connection.fetch("SELECT id, total FROM orders")

        with path.open("w", encoding="utf-8") as output:
            for row in rows:
                output.write(f"{row.id},{row.total}\n")
```

离开代码块时，`__exit__` 或 `__aexit__` 负责清理，即使发生异常也会执行。

需要注意：

- CPython 常能通过引用计数很快销毁对象，但不能依赖这一点管理外部资源；
- 循环引用、其他 Python 实现和解释器退出会改变销毁时机；
- 文件、连接和锁始终使用上下文管理器；
- 后台 asyncio Task 必须保存引用、支持取消并等待完成。

---

## 7. 资源应该活多久

| 生命周期 | 典型资源 | 谁创建和关闭 |
|---|---|---|
| 一次函数调用 | 临时文件、查询结果、锁 | 当前函数/作用域 |
| 一次请求 | 事务、请求级缓冲区 | Handler 或 application 用例 |
| 整个应用 | 数据库连接池、HTTP client、线程池 | main/bootstrap |
| 后台任务 | goroutine、Task、Worker、消费者 | 启动它的组件 |

常见错误是每个请求创建连接池，或者让一次请求创建的事务逃逸到全局对象。

关闭应用时通常按这个顺序：

```text
停止接收新请求
→ 发出取消信号
→ 等待正在执行的任务
→ 刷新日志和指标
→ 关闭连接池、文件和网络句柄
```

---

## 8. 项目中只问这 6 个问题

1. 谁创建这个资源？
2. 谁拥有并负责关闭它？
3. 正常、错误、取消三条路径都会关闭吗？
4. 资源应该活一次调用、一次请求，还是整个应用？
5. 后台任务能否取消并等待？
6. 内存增长来自语言堆，还是 Buffer、native 库等堆外资源？

---

## 语言内详细学习

- [C++：内存与资源](../cpp/08-内存与资源.md)
- [Go：内存与资源](../go/08-内存与资源.md)
- [Java：内存与资源](../java/08-内存与资源.md)
- [Node.js / TypeScript：内存与资源](../nodejs/08-内存与资源.md)
- [Python：内存与资源](../python/08-内存与资源.md)
