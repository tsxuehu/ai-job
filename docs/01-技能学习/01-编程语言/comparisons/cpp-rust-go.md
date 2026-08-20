# C++、Rust 与 Go：按程序员认知分类对比

> 学习阶段：L0—L2
>
> 比较基线：现代 C++、stable Rust、项目固定版本的 Go
>
> 最后复核：2026-08-20

本文不是罗列关键字，而是用程序员理解一门语言时真正关心的十二类问题比较 C++、Rust 和 Go。每一类都先看语义差异，再看同一任务如何表达。

## 先建立总体印象

| 语言 | 默认思维 | 最突出的能力 | 需要主动承担的复杂度 |
| --- | --- | --- | --- |
| C++ | 对象由谁拥有，生命周期何时结束，这个抽象成本多少 | 内存布局、硬件、ABI、模板、原生生态和极致控制 | 语言/构建复杂，允许悬垂、越界、data race 和其他未定义行为 |
| Rust | 值由谁拥有，当前是共享借用还是独占借用，失败是否进入类型 | 无 GC 的内存安全、所有权、enum/trait、Send/Sync | 所有权与生命周期建模、编译约束和 async 生态选择 |
| Go | 值复制后是否仍共享底层数据，goroutine 如何结束，容量是否有上限 | 小语言、GC、goroutine、网络标准库和统一工具链 | nil、GC、goroutine/连接泄漏、无界并发和运行时容量 |

三门语言都能实现网络服务和基础组件。差异主要是：哪些错误由语言阻止，哪些成本由运行时承担，哪些复杂度交给开发者和团队工具链。

## 1. 基础语法

### 规则对比

| 问题 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 源文件入口 | `main`，头文件/源文件/翻译单元 | `main`，crate/module | `package main` + `main` |
| 语句与块 | 分号 + 花括号 | 分号区分语句和块尾表达式 | 花括号，分号通常自动插入 |
| 名称空间 | namespace | module/crate 路径 | package |
| 元编程入口 | 预处理宏、模板、constexpr | 声明式/过程宏、泛型、const | 语言无宏；泛型、代码生成工具 |
| 可见性 | public/private 等及链接规则 | 默认私有，`pub` 精确开放 | 首字母大写导出 |
| 编译产物 | 通常原生机器码，链接模型复杂 | 通常原生机器码，由 Cargo 组织 | 原生机器码，go tool 统一构建 |

### 程序员需要记住

- C++ 的 `#include` 是预处理文本引入，还要理解声明、定义、翻译单元和链接。
- Rust 的块可以产生值，末尾多一个分号可能把结果变成 `()`；编译诊断经常直接描述所有权问题。
- Go 刻意减少语法选择，格式和花括号位置交给 `gofmt`，未使用名称与 import 会编译失败。

## 2. 变量、数据与类型

### 基本语义

| 问题 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 类型系统 | 静态类型，隐式转换相对丰富 | 静态强类型，转换倾向显式 | 静态类型，转换通常显式 |
| 类型推断 | `auto` | `let` | `:=` |
| 默认可变性 | 默认可变，`const` 限制 | 默认不可变，`mut` 开放 | 默认可变 |
| 未初始化 | 局部基础对象可能未初始化 | 使用前必须初始化 | 所有变量有零值 |
| 空值 | `nullptr`，或 `optional<T>` | 安全引用非空，`Option<T>` | pointer/slice/map/channel/interface 等可 nil |
| 文本 | `string` 拥有字节，`string_view` 借用 | `String` 拥有 UTF-8，`&str` 借用 | string 是不可变字节序列，通常为 UTF-8 |

### 同一任务：复制一个含动态序列的 Task

#### C++：拥有型容器默认值复制，也可移动

```cpp
Task first{"T-1", {"backend"}};
Task second = first;            // string/vector 深层值复制
Task third = std::move(first);  // 转移资源；first 仍可析构
```

#### Rust：非 Copy 值默认移动，深复制显式 clone

```rust
let first = Task {
    id: "T-1".to_owned(),
    tags: vec!["backend".to_owned()],
};
let second = first;          // move，first 不再可用
let third = second.clone();  // 显式深复制
```

#### Go：复制 struct 值，slice 仍共享底层数组

```go
first := Task{ID: "T-1", Tags: []string{"backend"}}
second := first
second.Tags[0] = "go" // first.Tags[0] 也改变
```

### 核心结论

`std::vector<T>`、`Vec<T>` 和 `[]T` 都像动态数组，但所有权语义完全不同：前两个通常拥有元素存储，Go slice 只是底层数组区间的描述值。理解赋值语义，比记住声明语法重要。

## 3. 表达式与逻辑控制

### 控制能力

| 能力 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 条件 | `if`、三元表达式 | `if` 本身产生值 | `if`，无三元表达式 |
| 多分支 | `switch`，传统分支可能贯穿 | `match` 模式解构且必须穷举 | `switch` 默认不贯穿 |
| 循环 | for/while/do/range-for | loop/while/for | 统一使用 for |
| 短路 | `&&`、`||` | `&&`、`||`，条件必须 bool | `&&`、`||`，条件必须 bool |
| 模式解构 | 有限，依赖结构化绑定/variant 访问 | 原生模式贯穿 let/match/if let | type switch、多返回值，模式能力较少 |

### 同一任务：处理任务状态

```cpp
// C++：enum + switch；新增状态需要团队保证完整处理
switch (state) {
case State::ready: return start();
case State::stopped: return Result::ok();
}
throw std::logic_error("unhandled state");
```

```rust
// Rust：带数据 enum + match；编译器检查穷举
match state {
    State::Ready { id } => start(id),
    State::Stopped => Ok(()),
}
```

```go
// Go：具名常量 + switch；通常保留明确 default 错误
switch state {
case StateReady:
    return start()
case StateStopped:
    return nil
default:
    return fmt.Errorf("unknown state: %v", state)
}
```

Rust 对封闭状态的表达最强；C++ 也能借助 `variant`/访问器增强；Go 保持直接，但完整性通常由测试和 default 分支保证。

## 4. 函数与作用域

### 参数语义

| 目的 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 读取大对象 | `const T&` | `&T` | 按值；大 struct 可用 `*T` |
| 修改调用方对象 | `T&` | `&mut T`，借用期独占 | `*T` 或共享描述值 |
| 转移所有权 | 按值 + move 约定 | 按值默认 move 非 Copy 值 | 永远复制变量值，底层可能共享 |
| 可空参数 | pointer/optional | `Option<&T>` 等显式表示 | pointer/interface 等 nil |
| 闭包 | lambda 显式/默认捕获 | 推导 Fn/FnMut/FnOnce | function literal 捕获外层变量 |
| 重载 | 支持 | 不支持传统重载 | 不支持 |

### 同一任务：修改标签

```cpp
bool add_tag(Task& task, std::string_view tag);
```

```rust
fn add_tag(task: &mut Task, tag: &str) -> Result<(), AddTagError>;
```

```go
func addTag(task *Task, tag string) error
```

三个签名都表达修改，但保证不同：C++ 引用通常非空却不能完整证明别名与寿命；Rust 可变借用在 safe code 中保证当前独占；Go pointer 可 nil、可被多个 goroutine 同时持有，调用约定和同步由开发者负责。

作用域也不等同于生命周期：C++ 局部 RAII 对象随作用域析构；Rust 所有者离开作用域 Drop，但借用可在最后使用后提前结束；Go 局部名称离开作用域后，对象只要仍可达就继续存活。

## 5. 自定义类型与抽象

### 面向对象能力不是“有没有 class”

| 概念 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 数据建模 | struct/class/enum/variant | struct/带数据 enum | struct/具名类型 |
| 封装 | private/public、构造与成员函数 | 模块可见性、私有字段、impl | package 可见性、未导出字段、method |
| 实现继承 | class 继承 | 不支持 | 不支持；embedding 是组合与提升 |
| 运行时多态 | 基类 + virtual | `dyn Trait` | interface |
| 编译期多态 | template + Concept | 泛型 + trait bound | 泛型 + constraint |
| 行为复用 | 继承、组合、模板 | 组合、默认 trait 方法、泛型 | 组合、embedding、普通函数 |

Rust 和 Go 不提供传统类继承，是因为它把三件经常被 class 继承绑定的事拆开：

1. 数据与实现复用使用组合。
2. 行为契约使用 trait/interface。
3. 运行时替换使用 trait object/interface value。

它们仍然支持封装和多态，只是不建立“继承字段与实现 = 子类型”的统一层级。

### 同一任务：存储抽象

```cpp
class TaskStore {
public:
    virtual ~TaskStore() = default;
    virtual std::optional<Task> find(std::string_view id) const = 0;
};
```

```rust
trait TaskStore {
    fn find(&self, id: &str) -> Result<Option<Task>, StoreError>;
}
// S: TaskStore 静态分派；&dyn TaskStore 动态分派
```

```go
type TaskStore interface {
    Find(context.Context, string) (Task, error)
}
// 方法集匹配即实现，通常由调用方定义小接口
```

C++ 的抽象选择最多，也最容易组合出过高复杂度；Rust trait 同时服务静态和动态分派；Go interface 结构化、隐式实现，鼓励小接口。

## 6. 容器与迭代

### 常用对应关系

| 需求 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 动态序列 | `vector<T>` | `Vec<T>` | `[]T` |
| 双端队列 | `deque<T>` | `VecDeque<T>` | 通常自行用 slice/ring 实现 |
| 哈希映射 | `unordered_map<K,V>` | `HashMap<K,V>` | `map[K]V` |
| 有序映射 | `map<K,V>` | `BTreeMap<K,V>` | 无内置；排序键或第三方结构 |
| 集合 | set/unordered_set | BTreeSet/HashSet | 常用 `map[T]struct{}` |
| 迭代 | iterator/ranges | Iterator trait/适配器 | range/普通循环 |

### 同一任务：筛选 ready id

```cpp
for (const auto& task : tasks) {
    if (task.ready()) ids.push_back(task.id());
}
```

```rust
let ids: Vec<_> = tasks.iter()
    .filter(|task| task.ready())
    .map(Task::id)
    .collect();
```

```go
ids := make([]TaskID, 0, len(tasks))
for _, task := range tasks {
    if task.Ready() { ids = append(ids, task.ID) }
}
```

关键差异不在语法长短：C++ 要掌握扩容和修改导致的迭代器失效；Rust 用借用检查拒绝许多遍历时破坏容器的写法；Go 要理解 slice 的长度、容量、append 重新分配和共享底层数组。

## 7. 错误处理

### 错误模型

| 问题 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 可恢复失败 | 异常、错误码、结果类型 | `Result<T,E>` | `(T, error)` |
| 不存在 | pointer/`optional<T>` | `Option<T>` | 零值 + bool、nil 或 error |
| 快速传播 | 异常展开或项目宏/结果设施 | `?` | `if err != nil { return ... }` |
| 不变量破坏 | assert/terminate/异常边界 | panic/abort | panic/recover 边界 |
| 错误是否进签名 | 取决于返回类型；异常通常不体现 | Result 明确体现错误类型 | error 体现会失败，不体现静态错误集合 |

### 同一任务：读取配置

```cpp
std::string read_config(const std::filesystem::path& path) {
    std::ifstream input(path);
    if (!input) throw ConfigError{"open failed"};
    return read_all(input);
}
```

```rust
fn read_config(path: &Path) -> Result<String, ConfigError> {
    std::fs::read_to_string(path).map_err(ConfigError::Read)
}
```

```go
func readConfig(path string) (string, error) {
    data, err := os.ReadFile(path)
    if err != nil { return "", fmt.Errorf("read config: %w", err) }
    return string(data), nil
}
```

C++ 项目必须先统一“异常还是显式结果”；Rust 用类型迫使调用方处理 Result；Go 用统一 error 接口保持直接，但错误类别、包装和重复日志需要团队规范。

## 8. 内存与资源

### 所有权和释放

| 问题 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 普通内存 | 值对象、RAII、智能指针，底层可手动 | ownership/borrow/Drop，unsafe 可越界 | GC 管理可达对象 |
| 确定性清理 | 析构函数 | Drop | 显式 Close + defer；GC 不保证及时 |
| 共享所有权 | `shared_ptr/weak_ptr` | `Rc/Arc/Weak` | GC 可达性；共享对象普通引用 |
| 悬垂/use-after-free | 语言允许，依赖抽象和工具 | safe Rust 阻止大量此类问题 | GC 避免对象过早释放 |
| 逻辑泄漏 | shared_ptr 环、容器、线程等 | Rc/Arc 环、任务、集合等 | 缓存、map、goroutine、timer、连接等 |

### 资源不是只有内存

- C++ 用 RAII 对象统一管理内存、文件、锁、连接和线程，异常路径也能清理。
- Rust 用所有权与 Drop 达到同样的确定性；引用计数只解决活多久，不解决谁能修改。
- Go 的 GC 不会替你关闭文件、HTTP body、数据库 rows、事务和 goroutine；成功获取后尽快 `defer Close`，并检查必要的关闭错误。

典型工具分别是：C++ 使用 ASan/UBSan/泄漏检测，Rust 使用类型系统并以 Miri/sanitizer 验证 unsafe，Go 使用 heap/goroutine profile 和资源指标定位可达对象与任务泄漏。

## 9. 模块化、依赖与 IO

### 工程边界

| 方面 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 代码组织 | header/source、namespace、library/module | package/crate/module/workspace | module/package/internal/cmd |
| 构建 | 编译器 + CMake 等 | Cargo | go tool |
| 依赖 | Conan/vcpkg/团队方案 | Cargo.toml/Cargo.lock | go.mod/go.sum |
| 主要复杂点 | 翻译单元、链接、ABI、平台和编译选项 | feature、编译时间、目标与 async 生态 | module、cgo、runtime/目标平台边界 |
| 网络 IO | 通常选择第三方库 | 同步标准库；异步依赖生态 | 标准库 HTTP/network 能力完整 |

### IO 共同要求

三门语言都必须处理部分读写、编码、协议版本、输入上限、超时、取消和背压。区别只是抽象入口：C++ 常由网络/序列化库定义，Rust 同步 `Read/Write` 或 async trait 生态，Go 以小型 `io.Reader/io.Writer` 接口组合。

依赖治理也有共同底线：固定工具链与版本、审查传递依赖和许可证、生成可复现产物、升级后验证兼容和性能。C++ 还必须格外关注编译器/运行库/ABI 组合。

## 10. 并发与异步

### 执行模型

| 问题 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 同步执行单元 | OS 线程 | OS 线程 | goroutine，由 runtime 调度到 OS 线程 |
| 共享状态 | mutex/atomic/condition variable | Mutex/atomic + ownership | mutex/atomic 或 channel |
| 编译期竞争防护 | 不阻止 data race | safe Rust 通过 Send/Sync 和借用阻止大量 data race | 不阻止 data race |
| 异步机制 | coroutine 是语言机制，runtime 由库提供 | Future/async-await，executor 由生态提供 | goroutine/channel/netpoll 内置 runtime |
| 取消 | stop token 或框架协议 | Future drop/token/runtime 协议 | Context 协作传播 |
| 竞争检测 | TSan | 编译检查 + 工具补充 | race detector |

### 谁负责停止 Worker

```cpp
std::jthread worker([](std::stop_token stop) {
    while (!stop.stop_requested()) run_one_task();
});
```

```rust
let worker = tokio::spawn(async move {
    tokio::select! {
        _ = token.cancelled() => Ok(()),
        result = run_worker() => result,
    }
});
```

```go
group, ctx := errgroup.WithContext(ctx)
group.Go(func() error { return worker(ctx) })
return group.Wait()
```

C++ 和 Rust 的语言协程/async 都不自带完整业务 runtime；Go 从调度、网络轮询到 goroutine 都由统一 runtime 提供。三者都不会自动提供背压、事务、幂等和优雅停机：队列必须有容量，每个任务必须有人等待或观察错误。

## 11. 运行时与性能

### 成本模型

| 维度 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 机器码 | 通常提前编译 | 通常提前编译 | 提前编译 |
| 语言级 GC | 无 | 无 | 有 |
| 普通对象释放 | 所有权/析构确定 | 所有权/Drop 确定 | GC 决定回收时机 |
| 泛型 | 模板实例化 | 通常单态化 | 由编译器实现，语义由类型参数约束 |
| 动态分派 | virtual/函数指针/类型擦除 | trait object | interface |
| 调度器 | 标准不提供统一任务 runtime | 标准不提供 async executor | runtime 提供 goroutine 调度与 netpoll |
| 常见尾延迟 | 分配、锁、系统调用、缓存/调度 | 分配、锁、执行器阻塞、系统调用 | GC、调度、分配、锁、下游排队 |

### 性能判断

- C++ 控制能力最细，但未定义行为、数据布局和构建配置也让优化更专业。
- Rust 无 GC 且类型系统提供强安全基线；clone、分配、锁、单态化体积和 async 状态机仍有成本。
- Go 以 GC 和统一调度换开发简洁；要持续观察分配率、heap、GC、goroutine、mutex/block 和连接池。

统一方法是：固定环境与业务负载 → 建立吞吐/P95/P99/CPU/内存基线 → profiler 定位热点 → 单变量修改 → 正确性回归。没有数据时，不优先使用对象池、无锁结构、unsafe 或定制 allocator。

## 12. 测试与工程实践

### 工具链对比

| 目标 | C++ | Rust | Go |
| --- | --- | --- | --- |
| 格式 | clang-format | rustfmt | gofmt |
| 静态检查 | 编译警告、clang-tidy 等 | 编译器、Clippy | go vet 等 |
| 测试 | GoogleTest/Catch2 等生态 | cargo test 内置工作流 | go test 内置工作流 |
| 内存/UB | ASan、UBSan、Valgrind 等 | safe 基线 + Miri/sanitizer | GC；heap/RSS/profile |
| 并发 | TSan、调试器 | Send/Sync + 工具 | race、mutex/block profile、trace |
| 性能 | perf/VTune/profiler/benchmark | criterion/系统 profiler/火焰图 | benchmark/pprof/trace |

C++ 工具能力强但需团队拼装构建矩阵；Rust Cargo 把格式、检查、测试、依赖集中起来；Go 的标准工具和网络库最统一。三者的 CI 都应覆盖单元、组件、集成、并发、取消、故障和不可信输入，并保存可诊断的构建版本。

生产服务的共同检查项：

- 请求、连接、线程/任务、队列、缓存和响应体有容量上限。
- 外部调用有连接/读取/总超时、取消和重试预算。
- 后台任务可停止、可等待、错误可观察。
- 指标覆盖流量、错误、延迟、饱和度与语言运行时资源。
- 支持过载拒绝、优雅停机、构建追溯和故障复盘。

## 如何选择

| 约束 | 优先评估 | 主要原因 |
| --- | --- | --- |
| 大量复用 C/C++ 库、驱动、硬件、原生 SDK | C++ | 生态、ABI 和既有投入 |
| 新建无 GC 且重视内存安全的系统组件 | Rust | 所有权、safe Rust、Send/Sync |
| 云原生控制面、网关、微服务和内部平台 | Go | 标准库、goroutine、统一工具链 |
| 极低延迟、精细布局、SIMD/allocator 控制 | C++ 或 Rust | 无语言级 GC，能控制布局与分配 |
| 团队希望快速交付并减少语言/构建选择 | Go | 语言与工具约定少 |
| 既要使用 C API，又希望新模块加强安全边界 | Rust FFI 或 C++ RAII 封装 | 取决于 ABI、边界规模和团队经验 |

选型还要综合团队经验、依赖生态、目标平台、交付期限、延迟、内存和长期维护。不存在脱离约束的“最佳语言”。

## 同项目练习

选择两门语言实现同一个有界任务服务：

1. `POST /tasks` 提交任务，队列满时拒绝。
2. Worker 支持超时、取消、错误传播和优雅停机。
3. 状态持久化，重复请求保持幂等。
4. 记录吞吐、P95/P99、CPU、内存、队列和失败。
5. 注入竞争、阻塞、资源未释放和任务泄漏问题。

对比时回答：哪些非法状态在编译期被拒绝？数据什么时候复制或共享？文件、连接、锁、任务何时释放？谁创建、取消并等待任务？哪类问题只能在运行期用工具发现？

## 最终检查

- [ ] 能解释 C++ 复制/移动、Rust move/borrow、Go 值复制与底层共享。
- [ ] 能解释三门语言如何实现封装、继承/组合、运行时与编译期多态。
- [ ] 能解释 RAII、Drop、GC 与文件/连接关闭的关系。
- [ ] 能解释 C++ coroutine、Rust Future/executor、Go goroutine/runtime。
- [ ] 能分别说出三门语言最常见的内存、并发、运行时和工程风险。
- [ ] 能根据具体项目约束完成选型，而不是只比较语法偏好。

继续学习：[C++](../cpp/README.md)、[Rust](../rust/README.md)、[Go](../go/README.md)。
