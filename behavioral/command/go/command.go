package main

import (
	"fmt"
	"strings"
)

// 场景：文本编辑器的撤销 / 重做（Undo / Redo）
//
// 痛点对比：
//   不用命令模式 → 撤销逻辑写在 Editor 里，每次操作前手动复制字段做"快照"
//                   新增"替换/格式化"操作要在 Editor 里再加一套备份/恢复代码
//                   撤销栈由调用方各自维护，难以复用，无法扩展到宏录制和协同编辑
//
//   用命令模式  → 每个操作封装为独立 Command（Execute + Undo），自带逆操作
//                   Editor（Invoker）只负责压栈/弹栈，新增操作只加一个 struct
//                   Command 对象可序列化、排队、重放，零改动 Editor 即可扩展新能力

// Command 接口
type Command interface {
	Execute()
	Undo()
	Description() string
}

// Receiver：被操作的文本缓冲区
type TextBuffer struct {
	content strings.Builder
}

func (b *TextBuffer) Insert(text string) {
	b.content.WriteString(text)
}

func (b *TextBuffer) Delete(n int) {
	s := b.content.String()
	if n > len(s) {
		n = len(s)
	}
	b.content.Reset()
	b.content.WriteString(s[:len(s)-n])
}

func (b *TextBuffer) Content() string {
	return b.content.String()
}

// 具体命令 1：插入文本
type InsertCommand struct {
	buffer *TextBuffer
	text   string
}

func (c *InsertCommand) Execute()          { c.buffer.Insert(c.text) }
func (c *InsertCommand) Undo()             { c.buffer.Delete(len(c.text)) }
func (c *InsertCommand) Description() string { return fmt.Sprintf("插入 %q", c.text) }

// 具体命令 2：删除末尾 N 个字符
type DeleteCommand struct {
	buffer  *TextBuffer
	n       int
	deleted string // 保存被删内容，用于 Undo
}

func (c *DeleteCommand) Execute() {
	s := c.buffer.Content()
	if c.n > len(s) {
		c.n = len(s)
	}
	c.deleted = s[len(s)-c.n:]
	c.buffer.Delete(c.n)
}

func (c *DeleteCommand) Undo()               { c.buffer.Insert(c.deleted) }
func (c *DeleteCommand) Description() string { return fmt.Sprintf("删除末尾 %d 个字符", c.n) }

// Invoker：命令历史，负责执行、撤销、重做
type Editor struct {
	buffer   *TextBuffer
	history  []Command // 已执行的命令栈
	redoStack []Command // 被撤销的命令栈
}

func NewEditor() *Editor {
	return &Editor{buffer: &TextBuffer{}}
}

func (e *Editor) Do(cmd Command) {
	cmd.Execute()
	e.history = append(e.history, cmd)
	e.redoStack = nil // 新操作清空 redo 栈
	fmt.Printf("  执行: %-22s → %q\n", cmd.Description(), e.buffer.Content())
}

func (e *Editor) Undo() {
	if len(e.history) == 0 {
		fmt.Println("  没有可撤销的操作")
		return
	}
	last := e.history[len(e.history)-1]
	e.history = e.history[:len(e.history)-1]
	last.Undo()
	e.redoStack = append(e.redoStack, last)
	fmt.Printf("  撤销: %-22s → %q\n", last.Description(), e.buffer.Content())
}

func (e *Editor) Redo() {
	if len(e.redoStack) == 0 {
		fmt.Println("  没有可重做的操作")
		return
	}
	cmd := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]
	cmd.Execute()
	e.history = append(e.history, cmd)
	fmt.Printf("  重做: %-22s → %q\n", cmd.Description(), e.buffer.Content())
}

func main() {
	editor := NewEditor()
	buf := editor.buffer

	fmt.Println("=== 执行操作 ===")
	editor.Do(&InsertCommand{buffer: buf, text: "Hello"})
	editor.Do(&InsertCommand{buffer: buf, text: ", World"})
	editor.Do(&InsertCommand{buffer: buf, text: "!!!"})
	editor.Do(&DeleteCommand{buffer: buf, n: 3})

	fmt.Println("\n=== 撤销 3 步 ===")
	editor.Undo()
	editor.Undo()
	editor.Undo()

	fmt.Println("\n=== 重做 2 步 ===")
	editor.Redo()
	editor.Redo()

	fmt.Println("\n=== 插入新内容（清空 redo 栈）===")
	editor.Do(&InsertCommand{buffer: buf, text: "!"})
	editor.Redo() // redo 栈已空

	fmt.Println("\n最终内容:", editor.buffer.Content())
}
