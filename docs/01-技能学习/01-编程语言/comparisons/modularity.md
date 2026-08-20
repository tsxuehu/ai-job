# 六门语言的模块化设计对比

模块系统回答“代码怎样声明、导入和构建”；模块化设计回答“系统怎样切分，边界如何保持稳定”。会使用 package、crate 或 import，不代表系统已经模块化。

## 先区分四个层次

| 层次 | 回答的问题 | 常见例子 |
| --- | --- | --- |
| 命名空间 | 名称如何避免冲突、控制可见性？ | C++ namespace、Rust module、Java/Go package |
| 源码模块 | 哪些代码形成高内聚边界？ | feature package、crate、library、npm/Python package |
| 构建模块 | 哪部分可独立编译、测试、发布和依赖？ | CMake target、Cargo crate、Maven module、workspace package |
| 部署单元 | 哪部分能独立运行、扩缩容和失败？ | 进程、服务、Worker、函数 |

这四层不必一一对应。模块化单体通常包含多个源码/构建模块，但仍作为一个进程部署。不要因为代码分了目录，就过早拆成微服务。

## 一个合格模块的六个条件

1. **职责高内聚**：围绕业务能力或稳定技术能力组织，而不是堆放同类文件。
2. **公共 API 很小**：调用方只依赖明确入口，不导入内部实现。
3. **依赖方向单向**：外层依赖内层，基础设施实现端口，不形成环。
4. **数据所有权清楚**：模块维护自己的不变量，不让外部直接修改内部集合和存储模型。
5. **可独立验证**：能通过公开 API 完成单元、组件或契约测试。
6. **替换成本可控**：数据库、网络和第三方 SDK 被适配层隔离，不污染核心模型。

## 六门语言如何表达模块边界

| 语言 | 主要源码边界 | 强制隐藏方式 | 构建/发布边界 | 最常见的模块化风险 |
| --- | --- | --- | --- | --- |
| C++ | namespace、头文件、library | public/private header、PImpl、target include | CMake target、静态/动态库、模块 | include 泄漏、ODR、循环依赖、ABI 与宏配置 |
| Rust | module、crate | 默认私有、`pub(crate)`、re-export | package/crate/workspace | `pub` 过度、crate 过细、feature 组合、trait 放错边界 |
| Go | package、`internal` | 小写名称、internal 导入限制 | module/package/binary | package 按技术层堆放、接口过大、import cycle、utils |
| Java | package、JPMS、构建 module | package-private、exports、架构测试 | JAR、Maven/Gradle module | Spring 扫描掩盖依赖、共享 entity、循环 Bean、公共 common 模块 |
| Node.js/TS | ESM module、package、workspace | package exports、私有路径约定 | npm package、bundle、进程 | 深层导入、ESM/CJS 混用、循环 import、类型模块与运行时错位 |
| Python | module、package、distribution | 下划线约定、`__all__`、稳定 re-export | wheel/package、进程 | import 副作用、循环 import、动态越界、全局单例与环境差异 |

语言提供的可见性强度不同，但模块边界最终都要靠 API、依赖规则、测试和代码评审共同维护。

## 推荐的依赖方向

```text
HTTP / CLI / 消息入口
          │
          ▼
       application
          │
          ▼
         domain

infrastructure ──实现──> application/domain 定义的端口

composition root：创建具体实现并完成组装
```

- domain 不依赖 Web 框架、ORM、数据库 driver 和第三方 SDK。
- application 编排用例、事务和端口，不泄露具体基础设施类型。
- infrastructure 依赖核心端口并提供实现，不能反过来让核心导入实现。
- 入口负责协议转换；composition root 是少数“知道所有具体类型”的位置。

这不是要求所有项目机械套四层。小项目可以合并目录，但依赖方向和边界职责仍应存在。

## 同一个任务服务如何切模块

### C++：让 CMake target 成为架构边界

```text
task_domain        # 值对象、状态机，无网络/数据库依赖
task_application   # 用例和 Storage 接口，依赖 task_domain
task_postgres      # 实现 Storage，依赖 application + 数据库库
task_http          # HTTP DTO/handler，依赖 application
task_server        # composition root，组装所有 target
```

公共头文件只暴露稳定 API；实现头文件不加入 PUBLIC include。`target_link_libraries` 的 PRIVATE/PUBLIC 传播关系应与架构依赖一致。

### Rust：crate 是昂贵但清晰的构建边界

```text
crates/domain
crates/application
crates/postgres
crates/http
crates/server
```

小项目先用一个 library crate 内的私有 module，边界稳定、确需独立编译或复用时再拆 crate。用 re-export 提供窄 API，不把所有项都 `pub`。端口 trait 通常放在需要抽象的一侧。

### Go：package 围绕能力，而不是 MVC 文件类型

```text
cmd/server
internal/task             # 领域和用例
internal/task/postgres    # 存储适配
internal/task/httpapi     # 协议适配
internal/platform         # 少量稳定跨业务能力
```

接口由消费方定义，通常保持一到数个方法。Go 禁止 import cycle；不要用搬到 `common` 或全局注册表掩盖错误依赖方向。

### Java：优先 package by feature

```text
task/
├── domain
├── application
├── adapter/in/web
├── adapter/out/persistence
└── config
```

先用 package-private 建立模块化单体；需要更强构建隔离时再拆 Maven/Gradle module 或使用 JPMS。Spring Bean 能被注入不代表依赖方向合理，可用架构测试阻止跨模块内部访问。

### Node.js / TypeScript：运行时 import 才是真依赖

```text
src/modules/task/domain
src/modules/task/application
src/modules/task/adapters/http
src/modules/task/adapters/postgres
src/bootstrap
```

每个模块用单一 public entry 导出稳定 API，禁止其他模块深层 import。TypeScript `interface` 会被擦除，运行时依赖注入需要 token、工厂或具体对象。避免 barrel file 制造隐蔽循环依赖。

### Python：package 边界要控制 import 副作用

```text
src/app/task/domain.py
src/app/task/application.py
src/app/task/adapters/http.py
src/app/task/adapters/postgres.py
src/app/bootstrap.py
```

用 `__init__.py` 谨慎 re-export 公共 API，内部模块以下划线或明确约定隐藏。Protocol 可描述消费方端口；bootstrap 显式组装对象，不在 import 时创建连接、线程和全局 client。

## 接口应该定义在哪里

一般原则是“接口属于使用它的一方”：

- application 需要保存任务，就由 application/domain 定义最小 Storage 端口。
- Postgres、内存和远程实现依赖该端口，而不是让核心依赖某个数据库 SDK 接口。
- 接口只包含当前调用方需要的行为，不能把具体实现的全部方法照抄进去。

语言差异：C++/Java 常显式继承 interface；Rust 显式 `impl Trait for Type`；Go 隐式满足接口；Python Protocol 与 TypeScript interface 主要提供结构化静态检查。无论语法如何，依赖方向原则相同。

## 数据不能直接穿透模块

不要让同一个数据库 entity/ORM model 同时充当领域对象、HTTP DTO 和消息 schema：

```text
HTTP DTO ⇄ application command/result ⇄ domain model ⇄ persistence record
```

转换看似增加代码，却隔离了外部协议、数据库结构和领域不变量。边界处完成运行时校验、错误转换、版本兼容和敏感字段控制。

## 循环依赖意味着什么

出现 A → B → A 时，优先问：

1. A/B 是否属于同一个高内聚模块，应该合并？
2. 是否有一段共同概念应该上移到更稳定的核心？
3. 是否应由调用方定义端口，让实现反向依赖抽象？
4. 是否通过事件传递事实，而不是同步调用回去？

不要用全局 Service Locator、事件总线、反射、动态 import 或 `common` 大包隐藏所有环；这些方式可能把编译期依赖变成更难追踪的运行时依赖。

## 模块与微服务

满足以下条件前，优先模块化单体：

- 边界和数据所有权已经稳定。
- 确实需要独立部署、扩缩容、权限或故障隔离。
- 团队能承担网络超时、重试、幂等、观测、部署和数据一致性成本。

把内部函数调用改成 RPC 不会自动改善模块化；边界不清时只会形成分布式耦合。

## 测试模块边界

- domain：通过公开行为测试不变量和状态转换。
- application：使用内存端口实现验证用例和错误。
- adapter：使用真实协议/数据库验证序列化、SQL 与资源边界。
- 跨模块：契约测试验证公共 API，架构测试或依赖检查阻止越界 import。
- composition root：少量启动测试确保依赖能正确组装。

## 评审清单

- [ ] 这个模块围绕业务能力还是文件类型/工具名称组织？
- [ ] 公共 API 是否比内部实现小，并且有稳定语义？
- [ ] 核心是否导入了框架、ORM、网络 SDK 或具体实现？
- [ ] 接口是否由消费方定义，是否只含实际需要的方法？
- [ ] DTO、领域对象和持久化模型是否被无边界混用？
- [ ] 是否存在循环依赖、深层导入、common/utils 大包或全局注册表？
- [ ] 模块能否通过公开 API 独立测试？
- [ ] 拆分构建模块或微服务是否有明确收益，而不是只增加目录和部署？

完整仓库中的 apps、modules、contracts、migrations、tests、deploy、ops 和 docs 如何协作，参见 [企业级后端项目工程结构](../../../03-实战项目/企业级项目工程结构.md)。

继续学习：[C++](../cpp/09-模块化依赖与IO.md)、[Rust](../rust/09-模块化依赖与IO.md)、[Go](../go/09-模块化依赖与IO.md)、[Java](../java/09-模块化依赖与IO.md)、[Node.js/TypeScript](../nodejs/09-模块化依赖与IO.md)、[Python](../python/09-模块化依赖与IO.md)。
