# 五门语言的并发与异步：同时处理 100 个订单

> 目标：对比五门语言如何限制并发、设置超时、传播取消，并避免共享状态竞争。

统一需求：

```text
输入：100 个待支付订单
每个订单：调用一次远程支付接口
并发上限：10
总超时：3 秒
任意任务失败：取消其余任务
```

只会“启动任务”不够。生产代码必须同时回答：

- 最多同时运行多少个？
- 谁等待任务结束？
- 超时后怎样通知任务停止？
- 失败是否取消其他任务？
- 结果写入共享集合时是否安全？

---

## 1. 一张表先看结论

| 语言 | 主要并发单位 | 常见异步机制 | 取消方式 |
|---|---|---|---|
| C++ | OS 线程、线程池任务 | future、coroutine、框架事件循环 | `stop_token` 或项目取消令牌，必须协作 |
| Go | goroutine | channel、`context`、errgroup | 取消 context，调用链必须检查/传递 |
| Java | 平台线程、虚拟线程、Executor 任务 | Future、CompletableFuture | interrupt、Future cancel、超时，必须协作 |
| Node.js/TS | 事件循环上的 Promise；Worker Thread | Promise、Stream | AbortSignal，底层 API 必须支持 |
| Python | asyncio Task、线程/进程任务 | coroutine、TaskGroup | Task cancellation，代码必须正确传播 |

异步不等于并行：

- 异步解决“等待 IO 时不要阻塞当前执行单元”；
- 并行解决“多个 CPU 工作同时执行”。

---

## 2. C++：线程池 + 协作取消

C++ 标准库没有统一的通用线程池接口。企业项目通常使用项目线程池、网络框架执行器或协程运行时。

下面的 `ThreadPool` 表示项目提供的线程池：

```cpp
void pay_all(std::span<const Order> orders) {
    ThreadPool pool{10};
    std::stop_source stop_source;
    std::vector<std::future<void>> tasks;

    for (const auto& order : orders) {
        tasks.push_back(pool.submit([&, order] {
            if (stop_source.stop_requested()) {
                return;
            }
            payment_client.pay(order, stop_source.get_token());
        }));
    }

    try {
        wait_all_with_timeout(tasks, 3s);
    } catch (...) {
        stop_source.request_stop();
        wait_all(tasks);
        throw;
    }
}
```

必须注意：

- `request_stop()` 只是发出请求，不会强杀线程；
- HTTP client、循环和阻塞等待必须主动响应停止；
- lambda 捕获的引用必须活到所有任务结束；
- 线程池析构前必须等待任务，不能让任务访问已销毁对象；
- 共享容器需要锁、原子操作或每任务独立结果。

协程也不会自动提供线程、超时和并发上限，这些仍由执行器和异步 IO 框架决定。

---

## 3. Go：goroutine + context + errgroup

```go
func PayAll(
    parent context.Context,
    orders []Order,
) error {
    ctx, cancel := context.WithTimeout(parent, 3*time.Second)
    defer cancel()

    group, ctx := errgroup.WithContext(ctx)
    group.SetLimit(10)

    for _, order := range orders {
        order := order
        group.Go(func() error {
            return paymentClient.Pay(ctx, order)
        })
    }

    return group.Wait()
}
```

发生了什么：

- 每个订单启动一个 goroutine；
- `SetLimit(10)` 保证最多十个任务同时执行；
- 第一个任务返回错误时，group 的 context 被取消；
- 三秒到期时，context 同样被取消；
- `Wait` 等待所有已启动任务退出。

取消能否生效取决于 `paymentClient.Pay` 是否把 context 传给 HTTP 请求，并在长循环中检查它。

不要启动无人等待、无法取消的 goroutine。goroutine 很轻，但数量、连接和下游容量仍然有限。

---

## 4. Java：Executor 限制并发，Future 负责等待

```java
void payAll(List<Order> orders) throws Exception {
    try (var executor = Executors.newFixedThreadPool(10)) {
        var futures = orders.stream()
            .map(order -> executor.submit(() -> {
                paymentClient.pay(order);
                return null;
            }))
            .toList();

        long deadline = System.nanoTime() + Duration.ofSeconds(3).toNanos();

        try {
            for (var future : futures) {
                long remaining = deadline - System.nanoTime();
                future.get(remaining, TimeUnit.NANOSECONDS);
            }
        } catch (Exception error) {
            futures.forEach(future -> future.cancel(true));
            throw error;
        }
    }
}
```

需要注意：

- `cancel(true)` 通过 interrupt 请求停止，不会安全地强杀线程；
- HTTP client 和业务循环必须正确处理中断/超时；
- 不能为每个请求随意创建新的大型线程池；
- Executor 属于应用级资源时，应由启动代码统一创建和关闭；
- 虚拟线程适合大量阻塞 IO，但仍要限制数据库连接和下游并发。

线程数量变便宜，不代表外部支付系统可以承受无限请求。

---

## 5. Node.js / TypeScript：Promise 并发 + AbortSignal

```ts
async function payAll(orders: Order[]): Promise<void> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 3000);
  let nextIndex = 0;

  async function worker(): Promise<void> {
    while (true) {
      const index = nextIndex++;
      if (index >= orders.length) return;
      await paymentClient.pay(orders[index], {
        signal: controller.signal,
      });
    }
  }

  try {
    await Promise.all(
      Array.from({ length: 10 }, () => worker())
    );
  } catch (error) {
    controller.abort();
    throw error;
  } finally {
    clearTimeout(timer);
  }
}
```

这里不是创建 100 个 OS 线程，而是在事件循环上推进多个异步 IO。

需要注意：

- Promise 本身不能被强制取消，底层操作必须接收 AbortSignal；
- `Promise.all` 在一个任务失败时立即拒绝，但其他操作只有收到取消并配合才会停止；
- 不限制并发地对 10000 个订单执行 `Promise.all` 会压垮连接池和下游；
- CPU 密集计算会阻塞事件循环，应使用 Worker Thread、native 模块或独立服务；
- Stream 还要处理背压，不能只监听 `data` 无限制读取。

---

## 6. Python：TaskGroup + Semaphore + timeout

```python
async def pay_all(orders: list[Order]) -> None:
    semaphore = asyncio.Semaphore(10)

    async def pay_one(order: Order) -> None:
        async with semaphore:
            await payment_client.pay(order)

    async with asyncio.timeout(3):
        async with asyncio.TaskGroup() as group:
            for order in orders:
                group.create_task(pay_one(order))
```

发生了什么：

- TaskGroup 负责保存并等待所有 Task；
- Semaphore 将同时执行的支付调用限制为十个；
- 一个任务失败时，TaskGroup 取消其他任务；
- timeout 到期时取消整个代码块。

需要注意：

- 不要捕获 `CancelledError` 后吞掉，否则结构化取消会失效；
- 同步数据库或 HTTP SDK 会阻塞事件循环，应使用异步客户端或线程适配；
- CPU 密集工作通常使用进程池、native 扩展或独立服务；
- 创建 Task 后必须有明确所有者，不能让任务静默丢失。

---

## 7. CPU 工作和 IO 工作怎么选

| 语言 | 网络/数据库 IO | CPU 密集任务 |
|---|---|---|
| C++ | 异步 IO、线程池或协程框架 | 有限线程池，并行算法 |
| Go | goroutine + 支持 context 的客户端 | goroutine 会由运行时调度，但仍受 CPU 核数限制 |
| Java | 虚拟线程/平台线程 + 客户端超时 | 有限平台线程池、ForkJoin 或专用执行器 |
| Node.js | Promise + 事件循环 | Worker Thread、native 或外部服务 |
| Python | asyncio + 异步客户端 | 进程池、native 或外部服务 |

不要把“语法上是 async”当成不会阻塞。只要内部调用同步阻塞 SDK，事件循环仍会卡住。

---

## 8. 共享结果最容易产生数据竞争

假设多个任务同时执行：

```text
successCount++
results.append(orderID)
```

必须判断这些操作是否允许并发：

- C++：使用 mutex、atomic 或每任务独立结果；
- Go：使用 channel、mutex、atomic，运行 `go test -race`；
- Java：使用并发集合、锁、atomic，明确 happens-before；
- Node.js：同一事件循环内没有线程数据竞争，但异步交错仍会产生逻辑竞争；Worker 间共享内存需要原子同步；
- Python：asyncio Task 也会在 `await` 处交错；线程/进程共享另有同步要求。

最简单的设计通常是每个任务返回独立结果，由等待方统一汇总。

---

## 9. 项目中只问这 7 个问题

1. 这是 IO 并发还是 CPU 并行？
2. 最大并发数由谁限制？
3. 总 deadline 和单次超时分别是多少？
4. 第一个任务失败后，其他任务怎样收到取消？
5. 谁等待所有任务真正结束？
6. 是否存在共享可变状态和数据竞争？
7. 下游连接池与服务容量能否承受当前并发？

---

## 语言内详细学习

- [C++：并发与异步](../cpp/10-并发与异步.md)
- [Go：并发与异步](../go/10-并发与异步.md)
- [Java：并发与异步](../java/10-并发与异步.md)
- [Node.js / TypeScript：并发与异步](../nodejs/10-并发与异步.md)
- [Python：并发与异步](../python/10-并发与异步.md)
