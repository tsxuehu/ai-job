# 五门语言如何实现面向对象

面向对象不是“必须有 class 和 extends”，而是一组组织数据、行为和依赖关系的思想。不同语言可能用 class、struct、interface、Protocol、闭包或组合实现相同概念。

## 1. 先把概念分开

| 概念 | 真正含义 | 常见误解 |
| --- | --- | --- |
| 对象 | 同时承载状态、行为或身份的运行时值 | 对象必须由 class 创建 |
| 类 | 创建和描述一类对象的语言机制 | 没有 class 就不能面向对象 |
| 封装 | 隐藏表示、限制修改、维护不变量 | 把字段 private 后生成 getter/setter |
| 抽象 | 只暴露调用方需要的能力，隐藏实现细节 | 抽象等于抽象基类 |
| 继承 | 从已有类型获得实现或建立子类型关系 | 代码复用必须使用继承 |
| 多态 | 相同调用契约作用于多种具体实现 | 多态只能依靠虚函数 |
| 组合 | 一个对象拥有或委托给其他对象完成能力 | 组合只是把类写成字段 |

最重要的区分：**实现复用、子类型关系和动态分派是三件事**。传统类继承经常同时提供三者，Go 则刻意用不同机制分别表达。

## 2. 五门语言总览

| 能力 | C++ | Go | Java | Node.js/TS | Python |
| --- | --- | --- | --- | --- | --- |
| class | 有 | 无 | 有 | JS 运行时 class + TS 类型 | 有，class 本身也是对象 |
| 数据类型 | struct/class/enum/variant | struct/具名类型 | class/record/enum | object/class/type/interface | class/dataclass/Enum |
| 封装边界 | 类 + namespace/module | package 导出规则 | 类 + package/module | 模块、闭包、`#field`、TS 检查 | 模块与命名约定、property |
| 类继承 | 支持，可多继承 | 不支持 | class 单继承 | class 单继承原型链 | 支持多继承与 MRO |
| 接口契约 | 抽象基类/Concept | interface | interface | interface/type/对象形状 | Protocol/ABC/鸭子类型 |
| 运行时多态 | virtual/类型擦除 | interface value | interface/基类动态分派 | 方法覆盖/对象行为 | 鸭子类型/ABC/方法覆盖 |
| 编译期多态 | template/Concept/重载 | 泛型 + constraint | 泛型检查/重载 | 泛型/联合/结构类型 | 类型检查器中的 Protocol/泛型 |
| 封闭变体 | enum/variant | 手工常量/类型组合 | enum/sealed 类型 | 判别联合 | Enum/类层次 + match |
| 复用首选 | 组合，必要时继承/模板 | 组合、embedding、函数 | 组合、接口默认方法 | 组合、函数、对象委托 | 组合、函数、Mixin（谨慎） |

## 3. 对象和类并不是同一个概念

### C++、Java、TypeScript、Python

这些语言具有 class 语法，可以把字段、构造和方法放在同一个声明中。对象通常是 class 的实例，但仍要区分：

- 值对象：由属性决定相等性，例如 Money、TaskId。
- 实体对象：具有跨状态变化保持的身份，例如 Order。
- 服务对象：主要封装协作和行为，通常没有领域身份。
- 数据传输对象：只承载边界数据，不承担核心业务不变量。

### Go

Go 把 class 拆成：

```text
type / struct   → 数据形状
receiver method → 行为
package         → 可见性与封装
interface       → 行为契约
```

method 可以定义在同 package 的具名类型上。对象能力存在，但没有 class hierarchy。

## 4. 封装如何实现

封装的目标是让对象无法轻易进入非法状态，而不是隐藏每个字段。

| 语言 | 主要机制 | 项目注意 |
| --- | --- | --- |
| C++ | private/protected、构造函数、const 方法、模块边界 | 返回引用/视图时仍要保证生命周期和可变性 |
| Go | package 内小写名称、构造函数、method | 封装粒度主要是 package，不是单个 struct |
| Java | private/package/protected/public、构造与方法 | final 引用不代表对象不可变；避免 setter 破坏不变量 |
| Node.js/TS | 模块、闭包、运行时 `#field`、TS private | TS private 通常擦除，不能当安全边界 |
| Python | 模块 API、前导下划线、property、不可变数据类型 | 主要是约定而非权限系统，需要测试和评审 |

### 示例：有效端口值

```go
type Port struct {
	value uint16
}

func NewPort(value uint16) (Port, error) {
	if value == 0 {
		return Port{}, ErrZeroPort
	}
	return Port{value: value}, nil
}

func (p Port) Value() uint16 {
	return p.value
}
```

这里的封装来自：字段私有、只能通过校验构造、公开行为不会破坏约束。把字段换成 public 再增加 setter，并不是真正封装。

## 5. 继承到底解决什么

继承可能同时承担：

1. **实现复用**：派生类型获得父类已有代码。
2. **子类型**：派生类型能替换父类型出现的位置。
3. **运行时分派**：通过父类型引用调用派生实现。

这三者混在一起会产生脆弱基类、深继承树和不正确替换。

### 支持类继承的语言

- C++：支持多继承，需要处理虚析构、对象切片、菱形继承和 ABI。
- Java：class 单继承、interface 多实现；抽象基类可复用骨架。
- TypeScript/JavaScript：`extends` 建立原型链；TS 额外做静态检查。
- Python：支持多继承，MRO 决定方法查找，Mixin 需要协作式 `super()`。

### Go 如何替代

| 需求 | Go |
| --- | --- |
| 复用字段 | struct 字段或 embedding |
| 复用行为 | 普通函数、embedding、显式委托 |
| 行为契约 | interface |
| 运行时替换 | interface value |
| 编译期替换 | 泛型 + constraint |

Go embedding 会提升方法，但不会建立“外层类型是内层类型子类”的关系。

## 6. 多态不止一种

| 多态类型 | 含义 | 语言示例 |
| --- | --- | --- |
| 子类型多态 | 通过共同接口在运行时替换实现 | C++ virtual、Go/Java interface |
| 参数多态 | 算法对多个类型通用 | C++/Go/Java/TS/Python 泛型 |
| 特设多态 | 同一名称按类型选择不同实现 | C++/Java 重载、Python 运算协议 |
| 封闭多态 | 所有变体已知并逐一匹配 | TS 判别联合、Java sealed、C++ variant |

运行时多态适合插件、依赖注入和运行时选择；编译期多态适合通用算法与性能敏感路径；封闭多态适合业务状态已经确定的模型。不要把所有变化都设计成 interface。

## 7. 同一任务的五种实现

目标：OrderService 依赖一个 Notifier，生产环境发送真实通知，测试环境替换成内存实现。这展示抽象、多态、组合和依赖注入，不需要继承业务实现。

### C++：抽象基类 + 组合

```cpp
class Notifier {
public:
    virtual ~Notifier() = default;
    virtual void send(const Order& order) = 0;
};

class OrderService {
public:
    explicit OrderService(Notifier& notifier) : notifier_{notifier} {}
    void confirm(const Order& order) { notifier_.send(order); }

private:
    Notifier& notifier_; // 非拥有借用，调用方保证生命周期
};
```

需要运行时切换时使用虚函数；若实现类型在编译期已知，也可用模板/Concept。签名必须明确 Notifier 的所有权。

### Go：小 interface + 组合

```go
type Notifier interface {
	Send(context.Context, Order) error
}

type OrderService struct {
	notifier Notifier
}

func (s *OrderService) Confirm(ctx context.Context, order Order) error {
	return s.notifier.Send(ctx, order)
}
```

具体类型通过方法集隐式满足接口。接口应由 OrderService 所在的消费方定义，只保留它需要的 `Send`。

### Java：interface + 构造注入

```java
public interface Notifier {
    void send(Order order);
}

public final class OrderService {
    private final Notifier notifier;

    public OrderService(Notifier notifier) {
        this.notifier = Objects.requireNonNull(notifier);
    }

    public void confirm(Order order) {
        notifier.send(order);
    }
}
```

Spring 可以在 composition root 绑定实现，但核心不需要主动查询 ApplicationContext。接口多态与 class 实现继承是两条独立机制。

### Node.js / TypeScript：结构接口 + 运行时对象

```ts
interface Notifier {
  send(order: Order, signal: AbortSignal): Promise<void>;
}

class OrderService {
  constructor(private readonly notifier: Notifier) {}

  confirm(order: Order, signal: AbortSignal): Promise<void> {
    return this.notifier.send(order, signal);
  }
}
```

任何形状兼容的对象都能传入。Notifier interface 在运行时不存在，真正注入的是构造函数接收的对象；外部输入仍需 schema 校验。

### Python：Protocol + 鸭子类型

```python
class Notifier(Protocol):
    async def send(self, order: Order) -> None: ...


class OrderService:
    def __init__(self, notifier: Notifier) -> None:
        self._notifier = notifier

    async def confirm(self, order: Order) -> None:
        await self._notifier.send(order)
```

Protocol 帮助静态检查；运行时仍是鸭子类型调用。若需要运行时名义约束和部分共享实现，可使用 ABC，但普通依赖替换通常不需要深继承。

## 8. 组合为什么通常优先

组合让依赖关系显示在字段和构造函数中：

```text
OrderService has a Notifier
```

继承表达的是：

```text
EmailNotifier is a Notifier
```

前者负责组装能力，后者负责替换关系。项目中大量需求只是“使用另一个能力”，应该组合；只有确实满足 is-a 和行为契约时才建立子类型。

组合的收益：

- 能独立替换、测试和管理依赖生命周期。
- 不继承不需要的方法和状态。
- 不受父类内部实现变化影响。
- 可以在运行时组合多个小能力。

组合也不能无限套代理和转发层。一个抽象只有在存在真实的第二实现、稳定边界、隔离外部依赖或明确测试价值时才值得建立。

## 9. 领域建模中的选择

| 问题 | 推荐机制 |
| --- | --- |
| 一个对象必须维护有效状态 | 私有字段 + 校验构造 + 行为方法 |
| 多个已知且封闭的业务状态 | enum/variant/sealed/判别联合 |
| 多个可扩展基础设施实现 | interface/Protocol/抽象基类 |
| 通用容器和算法 | 泛型/模板 |
| 只是复用几行无状态逻辑 | 普通函数，不创建继承体系 |
| 需要组合日志、重试、缓存 | 装饰器/中间件/包装对象，注意顺序与错误语义 |
| 需要复用数据字段 | 组合值对象；不要为了字段复用建立 is-a |

面向对象不是要求把所有函数放进类。领域对象适合状态与行为结合；无状态转换、解析和算法完全可以是普通函数。

## 10. 高频错误

- 贫血模型：对象只有 getter/setter，业务规则全部散落在 service。
- 上帝对象：一个类/struct 拥有所有依赖和所有业务。
- 深继承树：修改父类影响大量不可见调用方。
- 错误子类型：派生类不能遵守基类契约，却为了复用代码继承。
- 接口爆炸：每个类都机械创建一个接口，没有替换边界。
- 框架驱动模型：领域对象依赖 ORM、HTTP 和 DI 容器生命周期。
- 把 Go 的组合写成手工继承树，大量转发却没有清晰能力边界。
- 把 TypeScript/Python 的静态提示误认为运行时校验。

## 11. 面试需要能回答

- 封装为什么不是 private + getter/setter？
- 继承、组合、子类型和代码复用有什么区别？
- 运行时多态与编译期多态分别有什么成本？
- Go embedding 为什么不是继承？
- Java/C++ 什么时候应该使用抽象基类，什么时候只用接口？
- TypeScript interface 和 Python Protocol 在运行时是否存在？
- 为什么依赖注入通常依赖组合，而不是继承？
- 业务状态为什么有时更适合 enum/sealed/判别联合，而不是子类树？

## 12. 验收清单

- [ ] 能把对象、类、封装、继承、多态、组合分别定义清楚。
- [ ] 能在五门语言中指出对应机制，而不是只寻找 class 关键字。
- [ ] 能区分运行时多态、编译期多态和封闭状态多态。
- [ ] 能用组合和小接口实现可替换依赖，并说明所有权/生命周期。
- [ ] 能识别一个不满足替换原则的错误继承关系。
- [ ] 能根据领域状态、基础设施扩展和性能约束选择抽象方式。

语言内学习：[C++](../cpp/05-面向对象与抽象.md)、[Go](../go/05-面向对象与抽象.md)、[Java](../java/05-面向对象与抽象.md)、[Node.js/TypeScript](../nodejs/05-面向对象与抽象.md)、[Python](../python/05-面向对象与抽象.md)。
