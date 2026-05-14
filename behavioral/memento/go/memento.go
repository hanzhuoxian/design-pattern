package main

import (
	"fmt"
	"strings"
)

// 场景：文本编辑器的撤销 / 重做（Undo / Redo）
//
// 痛点对比：
//   不用备忘录 → 外部调用方直接复制 Editor 的字段来保存历史
//     Editor 一旦新增内部字段（如选区范围、格式标记），所有历史记录代码都要跟着改
//     若让外部直接操作内部字段，彻底破坏封装，Editor 内部结构无法自由演进
//
//   用备忘录 → Editor 自己决定"快照打包哪些内容"，外部只持有不透明的 Snapshot
//     History（Caretaker）只负责入栈 / 出栈，完全不感知 Editor 内部结构
//     Editor 随意增减字段，History 和调用方代码零改动

// ── Memento：快照（封装 Editor 的内部状态，不对外暴露细节）────────────────────

type Snapshot struct {
	content string // unexported：外部只能持有，无法直接读写
	cursor  int
	label   string
}

func (s *Snapshot) Label() string { return s.label }

// ── Originator：文本编辑器 ────────────────────────────────────────────────────

type Editor struct {
	content string
	cursor  int
}

func (e *Editor) Type(text string) {
	e.content = e.content[:e.cursor] + text + e.content[e.cursor:]
	e.cursor += len(text)
}

func (e *Editor) Delete(n int) {
	if n > e.cursor {
		n = e.cursor
	}
	e.content = e.content[:e.cursor-n] + e.content[e.cursor:]
	e.cursor -= n
}

func (e *Editor) MoveCursor(pos int) {
	switch {
	case pos < 0:
		e.cursor = 0
	case pos > len(e.content):
		e.cursor = len(e.content)
	default:
		e.cursor = pos
	}
}

// Save 打包当前状态为快照（Editor 自己决定快照内容，外部不可见细节）
func (e *Editor) Save(label string) *Snapshot {
	return &Snapshot{content: e.content, cursor: e.cursor, label: label}
}

// Restore 只有 Editor 能解析快照的内部字段
func (e *Editor) Restore(s *Snapshot) {
	e.content = s.content
	e.cursor = s.cursor
}

// Show 在光标位置插入 | 便于可视化
func (e *Editor) Show() string {
	return fmt.Sprintf("%q", e.content[:e.cursor]+"|"+e.content[e.cursor:])
}

// ── Caretaker：历史记录（只管入栈 / 出栈，不解读快照内容）─────────────────────

type History struct {
	undoStack []*Snapshot
	redoStack []*Snapshot
}

func (h *History) Push(s *Snapshot) {
	h.undoStack = append(h.undoStack, s)
	h.redoStack = nil // 新操作清空 redo 栈
}

func (h *History) Undo(e *Editor) bool {
	if len(h.undoStack) == 0 {
		return false
	}
	h.redoStack = append(h.redoStack, e.Save("(redo-point)"))
	top := h.undoStack[len(h.undoStack)-1]
	h.undoStack = h.undoStack[:len(h.undoStack)-1]
	e.Restore(top)
	return true
}

func (h *History) Redo(e *Editor) bool {
	if len(h.redoStack) == 0 {
		return false
	}
	h.undoStack = append(h.undoStack, e.Save("(undo-point)"))
	top := h.redoStack[len(h.redoStack)-1]
	h.redoStack = h.redoStack[:len(h.redoStack)-1]
	e.Restore(top)
	return true
}

func (h *History) Status() string {
	labels := make([]string, len(h.undoStack))
	for i, s := range h.undoStack {
		labels[i] = s.Label()
	}
	undo := strings.Join(labels, "→")
	if undo == "" {
		undo = "空"
	}
	return fmt.Sprintf("undo栈[%s]  redo栈(%d项)", undo, len(h.redoStack))
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	editor := &Editor{}
	history := &History{}

	do := func(label string, op func()) {
		history.Push(editor.Save(label))
		op()
		fmt.Printf("  %-20s 内容: %-28s %s\n", "["+label+"]", editor.Show(), history.Status())
	}

	fmt.Println("━━ 连续编辑 ━━")
	do("输入 Hello", func() { editor.Type("Hello") })
	do("输入 World", func() { editor.Type(" World") })
	do("光标移到头部", func() { editor.MoveCursor(0) })
	do("插入 Dear ", func() { editor.Type("Dear ") })

	fmt.Println("\n━━ 撤销两步 ━━")
	for i := 0; i < 2; i++ {
		if history.Undo(editor) {
			fmt.Printf("  [Undo]               内容: %-28s %s\n", editor.Show(), history.Status())
		}
	}

	fmt.Println("\n━━ 重做一步 ━━")
	if history.Redo(editor) {
		fmt.Printf("  [Redo]               内容: %-28s %s\n", editor.Show(), history.Status())
	}

	fmt.Println("\n━━ 重做后继续输入（redo 栈被清空）━━")
	do("输入 Great ", func() { editor.Type("Great ") })
	if !history.Redo(editor) {
		fmt.Println("  [Redo]  redo 栈为空，无法重做（新操作已截断 redo 历史）")
	}

	fmt.Println("\n━━ 删除后撤销 ━━")
	do("删除5个字符", func() { editor.Delete(5) })
	fmt.Printf("  删除后：%s\n", editor.Show())
	if history.Undo(editor) {
		fmt.Printf("  [Undo]               内容: %-28s %s\n", editor.Show(), history.Status())
	}
}
