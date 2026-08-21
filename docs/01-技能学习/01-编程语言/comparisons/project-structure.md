# 五门语言的工程结构组织对比

> 工程结构不是把同一套目录翻译成五种语言，而是用每门语言真正有效的机制管理代码边界。

本文仍使用同一个项目：一个包含订单、库存、HTTP API 和数据库的后端服务。

先记住结论：

- 业务代码按订单、库存等能力组织；
- `main` / `bootstrap` 只负责创建对象、连接模块和管理进程；
- 数据库、Web 框架等实现依赖业务接口，业务核心不依赖具体框架；
- 小项目先保持简单，边界稳定后再拆构建模块。

---

## 1. 五门语言一眼对比

| 语言 | 小中型项目的推荐起点 | 主要代码边界 | 可运行入口 | 何时拆成多个构建单元 |
|---|---|---|---|---|
| C++ | `apps + libs` | CMake target、公共头文件 | `apps/api/main.cpp` | 需要控制编译依赖、复用库或稳定 ABI 时 |
| Go | `cmd + internal` | package、`internal` | `cmd/api/main.go` | 不同部分确实需要独立版本或发布时才拆 module |
| Java | 单模块内 package by feature | package、可见性 | `bootstrap/Application.java` | 业务边界稳定，需要编译级约束时拆 Gradle/Maven module |
| Node.js/TS | 单 package 的 `src/modules` | ES module、`exports`、lint | `src/main.ts` | 需要独立构建、发布或 CI 缓存时拆 workspace package |
| Python | `src layout + modules` | Python package、公开入口、lint | `entrypoints/api.py` | 需要独立发布、复用或隔离依赖时拆 distribution |

最重要的区别是：

- C++ 和 Java 常把“构建单元”作为强边界；
- Go 首先依赖 package 和导入规则，通常不急着拆多个 `go.mod`；
- Node.js/TypeScript 可以用 workspace package 强化边界，但拆太细会增加构建成本；
- Python 的源码边界较软，更依赖公开入口、类型检查、lint 和评审。

---

## 2. 同一个订单服务，各语言怎么摆

下面都只展示足够表达边界的目录，不列部署、监控等通用目录。

### 2.1 C++：围绕 CMake target 组织

```text
repository/
├── CMakeLists.txt
├── CMakePresets.json
├── apps/
│   └── api/
│       └── main.cpp
├── libs/
│   ├── order/
│   │   ├── CMakeLists.txt
│   │   ├── include/company/order/
│   │   │   └── order_service.h
│   │   ├── src/
│   │   │   ├── order_service.cpp
│   │   │   └── order_model.h
│   │   └── tests/
│   └── inventory/
└── adapters/
    ├── order_http/
    └── order_postgres/
```

怎么看这个结构：

- 一个 CMake target 表示一个主要依赖单元；
- `include/company/order` 是调用方可见的合同；
- `src` 是私有实现，其他 target 不应 include；
- `apps/api` 链接订单、HTTP、数据库 target 并完成组装；
- 模块附近放单元测试，跨模块测试放仓库级 `tests`。

C++ 不适合照搬 Java 的层层 package。真正有约束力的是 target、头文件搜索路径和链接关系。

### 2.2 Go：围绕 package 和 `internal` 组织

```text
repository/
├── go.mod
├── go.sum
├── cmd/
│   ├── api/main.go
│   └── worker/main.go
└── internal/
    ├── order/
    │   ├── model.go
    │   ├── service.go
    │   ├── ports.go
    │   ├── httpapi/
    │   └── postgres/
    ├── inventory/
    └── platform/
        ├── config/
        └── observability/
```

怎么看这个结构：

- `cmd/api` 是 API 进程入口，`cmd/worker` 是 Worker 入口；
- `internal/order` 是订单业务 package；
- `httpapi`、`postgres` 依赖订单业务，而不是订单业务导入它们；
- 接口由使用方定义，例如订单服务定义自己需要的 `Store`；
- 测试通常和被测 `.go` 文件放在同一 package。

Go 项目通常先使用一个 `go.mod`。不要为了“模块化”给每个业务目录创建一个 module，也不要机械创建 `controller/service/repository` 三个 package。

### 2.3 Java：先 package by feature，再考虑多模块

中小项目推荐先用一个构建模块：

```text
repository/
├── build.gradle.kts
└── src/
    ├── main/java/com/company/app/
    │   ├── order/
    │   │   ├── domain/
    │   │   ├── application/
    │   │   └── adapter/
    │   │       ├── web/
    │   │       └── persistence/
    │   ├── inventory/
    │   └── bootstrap/
    │       └── Application.java
    └── test/java/com/company/app/
```

怎么看这个结构：

- 第一层按业务分 `order`、`inventory`，不是全局 controller/service/repository；
- 业务模块内部再区分 domain、application 和 adapter；
- package-private 隐藏不需要公开的类；
- `bootstrap` 启动 Spring 并组装 Bean；
- 测试目录与生产 package 对应。

当边界已经稳定并且确实需要编译级限制时，再升级：

```text
modules/
├── order-domain/
├── order-application/
├── order-adapter-web/
└── order-adapter-persistence/
apps/
└── boot-api/
```

不要一开始就把每个 package 变成一个 Gradle/Maven module。模块越多，构建和依赖管理成本越高。

### 2.4 Node.js / TypeScript：先单 package，再升级 workspace

中小项目推荐：

```text
repository/
├── package.json
├── tsconfig.json
└── src/
    ├── main.ts
    ├── modules/
    │   ├── order/
    │   │   ├── index.ts
    │   │   ├── domain/
    │   │   ├── application/
    │   │   └── adapters/
    │   └── inventory/
    └── platform/
        ├── config/
        └── observability/
```

怎么看这个结构：

- `src/main.ts` 加载配置、创建连接并组装模块；
- 每个业务模块用 `index.ts` 选择公共 API；
- 其他模块只能从订单入口导入，禁止深层导入内部文件；
- 外部 JSON 在 adapter 做运行时校验，不能只靠 TypeScript 类型；
- 单元测试可贴近源码，集成测试放仓库级 `tests`。

需要多个应用、独立构建或发布时，再使用 workspace：

```text
apps/
├── api/
└── worker/
packages/
├── order-domain/
├── order-application/
└── order-adapter-postgres/
```

workspace package 必须配置明确的 `exports`。仅有 TypeScript path alias，不代表运行时边界真的存在。

### 2.5 Python：使用 src layout，显式设计 package API

```text
repository/
├── pyproject.toml
├── src/
│   └── company_app/
│       ├── modules/
│       │   ├── order/
│       │   │   ├── __init__.py
│       │   │   ├── domain.py
│       │   │   ├── application.py
│       │   │   ├── ports.py
│       │   │   └── adapters/
│       │   └── inventory/
│       ├── platform/
│       └── entrypoints/
│           ├── api.py
│           └── worker.py
└── tests/
    ├── unit/
    └── integration/
```

怎么看这个结构：

- `src layout` 防止测试意外导入仓库根目录里的源码；
- `modules/order/__init__.py` 只导出稳定的公共名字；
- `_name` 表示内部实现，但主要依靠团队约定和工具检查；
- `entrypoints` 创建 Web 应用、连接池和业务服务；
- 同步和异步 adapter 要分清，不能在事件循环里直接跑阻塞 IO。

Python 不要按文件类型创建全局 `models.py`、`services.py`、`utils.py`。项目变大后，这些文件会变成没有边界的杂物间。

---

## 3. 同一职责在五门语言中的位置

| 职责 | C++ | Go | Java | Node.js/TS | Python |
|---|---|---|---|---|---|
| API 进程入口 | `apps/api/main.cpp` | `cmd/api/main.go` | `bootstrap/Application.java` 或 `boot-api` | `src/main.ts` 或 `apps/api` | `entrypoints/api.py` |
| 订单业务模型 | `libs/order/src` | `internal/order` | `order/domain` | `modules/order/domain` | `modules/order/domain.py` |
| 订单公开入口 | 公共头文件 | 大写名字/公开 package API | `public` 类型或 API package | `index.ts` / `exports` | `__init__.py` / `__all__` |
| 数据库实现 | adapter target | `order/postgres` | `adapter/persistence` | `adapters/postgres` | `adapters/postgres.py` |
| 编译/构建边界 | CMake target | package；必要时 module | Gradle/Maven module | workspace package | distribution/package |
| 单元测试 | target 内 `tests` | 同 package `_test.go` | `src/test` 对应 package | 源码附近或 `tests/unit` | `tests/unit` 或源码对应目录 |
| 依赖组装 | app target | `cmd` | Spring boot/config | `main` / bootstrap | entrypoint/bootstrap |

目录名可以不同，但职责是一一对应的。

---

## 4. 哪些边界是语言真正能强制的

| 语言 | 编译器/运行时能直接阻止什么 | 仍需要工具或评审防止什么 |
|---|---|---|
| C++ | 私有成员、不可见头文件、未链接 target | 公共头文件泄露、错误的 PUBLIC 依赖、ABI 风险 |
| Go | 小写名字不可导出、`internal` 限制、循环 import | 大 package、`common` 滥用、业务间越界调用 |
| Java | private/package-private、构建依赖、JPMS exports | 所有类都 public、Spring/JPA 渗入 domain |
| Node.js/TS | package `exports` 可限制发布入口 | 仓库内深层导入、类型与运行时不一致、循环依赖 |
| Python | 几乎没有不可突破的源码权限墙 | 私有约定、依赖方向、动态导入和运行时类型边界 |

所以不能只看目录。真正的工程结构是：

```text
目录布局
+ 语言可见性
+ 构建依赖
+ 静态检查
+ 测试
+ 代码评审规则
```

---

## 5. 五门语言都应该遵守的依赖方向

```text
API / Worker 入口
        │
        ▼
HTTP、消息等输入 adapter
        │
        ▼
应用用例
        │
        ▼
业务模型
        ▲
        │ 实现业务所需接口
数据库、缓存、外部服务 adapter
```

启动入口是唯一可以同时知道抽象和具体实现的地方：

```text
main/bootstrap:
    创建数据库连接
    创建 OrderStore 实现
    创建 OrderService
    创建 HTTP Handler
    启动并在退出时关闭资源
```

业务规则不要写在 `main`、controller、route handler 或 ORM entity 中。

---

## 6. 项目从小到大怎么演进

| 阶段 | 推荐结构 | 不要急着做什么 |
|---|---|---|
| 练习/原型 | 一个应用入口 + 按业务分目录 | 不创建十几个空层次 |
| 中小型生产项目 | 模块化单体 + 单向依赖 + 分层测试 | 不按每个业务拆仓库或服务 |
| 边界稳定的单体 | 用 target/module/workspace 强化构建边界 | 不让公共模块变成垃圾场 |
| 多进程应用 | API、Worker、Scheduler 分入口，共享核心模块 | 不复制业务代码 |
| 独立服务 | 独立数据、部署、容量和团队责任 | 不只因为目录大就拆服务 |

语言对应的升级动作：

- C++：从少量 target 演进到清晰的 app/library/adapter target 图；
- Go：先拆 package，再按独立发布需求决定是否拆 module；
- Java：从 package by feature 演进到多 Gradle/Maven module；
- Node.js/TS：从单 package 演进到 workspace；
- Python：从单 distribution 内多个 package，演进到少量可独立发布的 distribution。

---

## 7. 最容易犯的“照搬”错误

### 把 Java 结构照搬到 Go

为每个业务创建 controller、service、repository package，结果接口和类型来回引用，迅速产生 import cycle。

### 把 Go 的扁平结构照搬到 Java

所有类都塞进一个 package 并全部 `public`，Spring、JPA 和业务规则混在一起。

### 把 Node.js workspace 当成目录整理工具

每几个文件建一个 package，最后构建、版本和循环依赖治理比业务本身还复杂。

### 把 Python 下划线当成强权限

调用方仍能导入内部实现，必须配合公开入口、lint、类型检查和评审。

### 只整理 C++ 目录，不整理 target

源码看起来分开了，却仍全部编译进一个 target，并共享全局 include path，实际上没有边界。

---

## 8. 面试时怎么回答

不要背目录树，可以这样说：

> 我会先选择模块化单体，按订单、库存等业务能力组织代码。每个模块公开少量入口，数据库和 Web 框架放在 adapter，main/bootstrap 负责依赖组装和资源生命周期。然后使用该语言真正有效的边界机制：C++ 的 CMake target 和公共头文件、Go 的 package/internal、Java 的 package 与构建模块、Node.js 的 exports/workspace、Python 的 package API 与工具检查。只有边界稳定且存在独立构建或发布收益时，才继续拆模块或服务。

这个回答同时说明了：

- 你知道目录只是表面；
- 你理解依赖方向；
- 你知道不同语言不能机械套模板；
- 你能控制工程复杂度。

---

## 9. 选择结构时检查这 6 件事

1. 一个订单需求主要修改订单目录，还是要跑遍整个仓库？
2. 其他模块能否绕过公开入口访问订单内部实现？
3. 构建工具能否看出并限制依赖关系？
4. 核心业务能否不启动 Web 框架和数据库直接测试？
5. API、Worker 的资源由谁创建、关闭和等待？
6. 新增目录或构建模块解决了什么真实问题？

能清楚回答这六个问题，工程结构才不是摆设。

---

## 详细语言模板

- [C++ 企业级项目结构](../cpp/企业级项目结构.md)
- [Go 企业级项目结构](../go/企业级项目结构.md)
- [Java 企业级项目结构](../java/企业级项目结构.md)
- [Node.js / TypeScript 企业级项目结构](../nodejs/企业级项目结构.md)
- [Python 企业级项目结构](../python/企业级项目结构.md)
- [通用企业级后端项目工程结构](../../../03-实战项目/企业级项目工程结构.md)
- [五门语言的模块化设计对比](modularity.md)
