# 设计模式

> 23 种经典设计模式的多语言实现参考，基于 GoF《设计模式：可复用面向对象软件的基础》

---

## 目录

- [设计模式](#设计模式)
  - [目录](#目录)
  - [简介](#简介)
  - [模式分类](#模式分类)
    - [创建型模式](#创建型模式)
    - [结构型模式](#结构型模式)
    - [行为型模式](#行为型模式)
  - [语言支持](#语言支持)
  - [目录结构](#目录结构)
  - [如何使用](#如何使用)
  - [贡献指南](#贡献指南)
  - [参考资料](#参考资料)
  - [许可证](#许可证)

---

## 简介

设计模式是软件开发中针对常见问题的通用、可复用解决方案。它们不是可以直接转化为代码的成品，而是描述如何解决特定上下文中反复出现问题的模板。

本仓库收录了 GoF（Gang of Four）提出的全部 23 种设计模式，并提供多种编程语言的实现示例，帮助开发者理解每种模式的核心思想与适用场景。

---

## 模式分类

### 创建型模式

> 关注对象的**创建机制**，将对象的创建与使用解耦。

| # | 模式 | 说明 | 详情 |
|---|------|------|------|
| 01 | 单例模式 Singleton | 确保一个类只有一个实例，并提供全局访问点 | [查看](./creational/singleton/) |
| 02 | 工厂方法模式 Factory Method | 定义创建对象的接口，由子类决定实例化哪个类 | [查看](./creational/factory_method/) |
| 03 | 抽象工厂模式 Abstract Factory | 提供创建一系列相关对象的接口，无需指定具体类 | [查看](./creational/abstract_factory/) |
| 04 | 建造者模式 Builder | 将复杂对象的构建与表示分离，同样的构建过程可创建不同表示 | [查看](./creational/builder/) |
| 05 | 原型模式 Prototype | 通过复制现有对象来创建新对象，而非重新实例化 | [查看](./creational/prototype/) |

### 结构型模式

> 关注类与对象的**组合方式**，形成更大的结构。

| # | 模式 | 说明 | 详情 |
|---|------|------|------|
| 06 | 适配器模式 Adapter | 将一个类的接口转换为客户端期望的另一个接口 | [查看](./structural/adapter/) |
| 07 | 桥接模式 Bridge | 将抽象部分与实现部分分离，使二者可独立变化 | [查看](./structural/bridge/) |
| 08 | 组合模式 Composite | 将对象组合成树形结构，以表示"部分-整体"层次结构 | [查看](./structural/composite/) |
| 09 | 装饰器模式 Decorator | 动态地为对象添加新功能，比继承更灵活 | [查看](./structural/decorator/) |
| 10 | 外观模式 Facade | 为子系统提供一个统一的高层接口 | [查看](./structural/facade/) |
| 11 | 享元模式 Flyweight | 通过共享大量细粒度对象来节省内存 | [查看](./structural/flyweight/) |
| 12 | 代理模式 Proxy | 为另一个对象提供代理或占位符，以控制对其的访问 | [查看](./structural/proxy/) |

### 行为型模式

> 关注对象之间的**职责分配**与**通信方式**。

| # | 模式 | 说明 | 详情 |
|---|------|------|------|
| 13 | 职责链模式 Chain of Responsibility | 将请求沿处理者链传递，直到某个处理者处理它 | [查看](./behavioral/chain_of_responsibility/) |
| 14 | 命令模式 Command | 将请求封装为对象，支持撤销、队列、日志等操作 | [查看](./behavioral/command/) |
| 15 | 迭代器模式 Iterator | 提供顺序访问集合元素的方法，无需暴露内部结构 | [查看](./behavioral/iterator/) |
| 16 | 中介者模式 Mediator | 用中介对象封装一系列对象的交互，降低耦合 | [查看](./behavioral/mediator/) |
| 17 | 备忘录模式 Memento | 捕获并外部化对象的内部状态，以便后续恢复 | [查看](./behavioral/memento/) |
| 18 | 观察者模式 Observer | 定义对象间一对多的依赖，状态变化时自动通知所有依赖者 | [查看](./behavioral/observer/) |
| 19 | 状态模式 State | 允许对象在内部状态改变时改变其行为 | [查看](./behavioral/state/) |
| 20 | 策略模式 Strategy | 定义一系列算法，将其封装并可相互替换 | [查看](./behavioral/strategy/) |
| 21 | 模板方法模式 Template Method | 在父类中定义算法骨架，子类覆盖特定步骤 | [查看](./behavioral/template_method/) |
| 22 | 访问者模式 Visitor | 在不改变元素类的前提下，为其添加新操作 | [查看](./behavioral/visitor/) |
| 23 | 解释器模式 Interpreter | 为语言定义文法，并提供解释器解释该语言中的句子 | [查看](./behavioral/interpreter/) |

## 语言支持

| 语言 | 状态 | 说明 |
|------|------|------|
| Go | ✅ 已实现 | 每个模式位于对应目录的 `go/` 子目录下 |
| Python | 📅 计划中 | - |
| Java | 📅 计划中 | - |
| TypeScript | 📅 计划中 | - |
| Rust | 📅 计划中 | - |

---

## 目录结构

每个模式以**模式为第一层、语言为第二层**组织，便于横向对比不同语言的实现：

```
design-pattern/
├── README.md
├── LICENSE
├── cmd/                             # 脚手架工具
│   ├── go.mod
│   └── designctl.go                 # 自动生成模式目录和文件骨架
│
├── creational/                      # 创建型模式
│   ├── README.md
│   ├── singleton/
│   │   ├── README.md
│   │   └── go/
│   ├── factory_method/
│   │   ├── README.md
│   │   └── go/
│   ├── abstract_factory/
│   │   ├── README.md
│   │   └── go/
│   ├── builder/
│   │   ├── README.md
│   │   └── go/
│   └── prototype/
│       ├── README.md
│       └── go/
│
├── structural/                      # 结构型模式
│   ├── README.md
│   └── [adapter/bridge/composite/decorator/facade/flyweight/proxy]/
│       ├── README.md
│       └── go/
│
└── behavioral/                      # 行为型模式
    ├── README.md
    └── [chain_of_responsibility/command/interpreter/iterator/
        mediator/memento/observer/state/strategy/template_method/visitor]/
        ├── README.md
        └── go/
```

每个模式目录下包含：
- `README.md`：模式说明、适用场景、结构图、优缺点
- `go/`：Go 语言实现（独立 Go module）

---

## 如何使用

每个模式的 Go 实现是独立的 Go module，进入对应目录运行：

```bash
# 以单例模式为例
cd creational/singleton/go
go test ./...
```

批量运行所有 Go 测试：

```bash
find . -name 'go.mod' -not -path './cmd/*' | while read mod; do
  dir=$(dirname "$mod")
  echo "==> $dir"
  (cd "$dir" && go test ./...)
done
```

使用脚手架工具为新模式生成目录和文件骨架：

```bash
cd cmd
go run designctl.go
```

---

## 贡献指南

欢迎贡献新的语言实现或改进现有代码！

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feat/add-rust-singleton`
3. 在对应模式目录下新建语言子目录，添加实现和测试
4. 确保代码通过测试
5. 提交 PR，描述你的改动

**实现规范**：
- 代码需包含完整的单元测试
- 遵循对应语言的惯用写法（idiomatic code）
- 避免引入不必要的外部依赖

---

## 参考资料

- 📖 《设计模式：可复用面向对象软件的基础》—— GoF（Erich Gamma 等）
- 📖 《Head First 设计模式》—— Freeman & Robson
- 🌐 [Refactoring.Guru 设计模式](https://refactoring.guru/design-patterns)

---

## 许可证

[MIT](./LICENSE)
