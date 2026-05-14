package main

import "fmt"

// 场景：跨平台桌面应用 UI 渲染（Windows / macOS）
//
// 痛点对比：
//   不用抽象工厂 → new(WindowsButton) + new(MacCheckbox) 编译通过，但运行期 UI 风格混乱
//                   新增 Linux 平台需要在每个业务调用点散改，极易遗漏导致控件跨平台混用
//
//   用抽象工厂  → 工厂负责"整族对象的创建"，同一工厂产出的控件天然风格一致
//                   新增平台只加一个 Factory 实现，业务代码（Application）零改动

// ---------- 抽象产品 ----------

type Button interface {
	Render()
}

type Checkbox interface {
	Render()
}

// ---------- 抽象工厂 ----------
// 一个工厂只生产同一平台的全套组件，保证家族内部兼容。

type UIFactory interface {
	CreateButton() Button
	CreateCheckbox() Checkbox
}

// ---------- Windows 家族 ----------

type WindowsButton struct{}

func (b *WindowsButton) Render() {
	fmt.Println("[Windows] 渲染方形按钮，蓝色边框")
}

type WindowsCheckbox struct{}

func (c *WindowsCheckbox) Render() {
	fmt.Println("[Windows] 渲染方形复选框，打勾样式")
}

type WindowsFactory struct{}

func (WindowsFactory) CreateButton() Button   { return &WindowsButton{} }
func (WindowsFactory) CreateCheckbox() Checkbox { return &WindowsCheckbox{} }

// ---------- macOS 家族 ----------

type MacButton struct{}

func (b *MacButton) Render() {
	fmt.Println("[macOS]   渲染圆角按钮，灰色渐变")
}

type MacCheckbox struct{}

func (c *MacCheckbox) Render() {
	fmt.Println("[macOS]   渲染圆形复选框，蓝色填充")
}

type MacFactory struct{}

func (MacFactory) CreateButton() Button   { return &MacButton{} }
func (MacFactory) CreateCheckbox() Checkbox { return &MacCheckbox{} }

// ---------- 客户端代码 ----------
// Application 只依赖抽象工厂，不知道也不关心具体平台。
// 它永远拿到一套风格一致的组件，不可能出现"Windows 按钮 + macOS 复选框"的混用。

type Application struct {
	button   Button
	checkbox Checkbox
}

func NewApplication(factory UIFactory) *Application {
	return &Application{
		button:   factory.CreateButton(),
		checkbox: factory.CreateCheckbox(),
	}
}

func (app *Application) RenderUI() {
	app.button.Render()
	app.checkbox.Render()
}

// ---------- 演示 ----------

func main() {
	// 运行时根据当前平台选择工厂，其余代码完全不变
	platforms := map[string]UIFactory{
		"Windows": WindowsFactory{},
		"macOS":   MacFactory{},
	}

	for name, factory := range platforms {
		fmt.Printf("=== %s 平台 ===\n", name)
		app := NewApplication(factory)
		app.RenderUI()
		fmt.Println()
	}
}
