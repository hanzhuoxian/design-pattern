// Composite 模式——以文件系统为例
//
// 问题：目录可以包含文件，也可以包含子目录。如果不用该模式，
// 计算大小、打印结构等操作都需要写 if file / if directory 的分支，
// 且每个使用方都要自己做递归——逻辑分散、难以扩展。
//
// 解决：File（叶）和 Directory（组合）实现同一接口 Node，
// 调用方对任意节点调用 Size() / Print()，无需关心内部结构。

package main

import "fmt"

// Node 统一接口，文件和目录都实现它
type Node interface {
	Size() int
	Print(indent string)
}

// File 是叶节点，没有子节点
type File struct {
	name string
	size int // KB
}

func (f *File) Size() int { return f.size }

func (f *File) Print(indent string) {
	fmt.Printf("%s[file] %s  %d KB\n", indent, f.name, f.size)
}

// Directory 是组合节点，可包含任意 Node（File 或嵌套 Directory）
type Directory struct {
	name     string
	children []Node
}

func (d *Directory) Add(n Node) {
	d.children = append(d.children, n)
}

// Size 递归求和——File.Size() 直接返回，Directory.Size() 继续递归
// 调用方不需要知道这里面有多少层嵌套
func (d *Directory) Size() int {
	total := 0
	for _, child := range d.children {
		total += child.Size()
	}
	return total
}

func (d *Directory) Print(indent string) {
	fmt.Printf("%s[dir]  %s  %d KB\n", indent, d.name, d.Size())
	for _, child := range d.children {
		child.Print(indent + "  ")
	}
}

func main() {
	// 构建目录树
	root := &Directory{name: "project"}

	src := &Directory{name: "src"}
	src.Add(&File{name: "main.go", size: 12})
	src.Add(&File{name: "handler.go", size: 34})

	images := &Directory{name: "images"}
	images.Add(&File{name: "logo.png", size: 200})
	images.Add(&File{name: "banner.jpg", size: 450})

	assets := &Directory{name: "assets"}
	assets.Add(images)
	assets.Add(&File{name: "style.css", size: 18})

	root.Add(src)
	root.Add(assets)
	root.Add(&File{name: "README.md", size: 5})

	// 对整棵树调用——和对单个文件调用的代码完全一样
	fmt.Println("=== 完整目录树 ===")
	root.Print("")
	fmt.Printf("项目总大小: %d KB\n", root.Size())

	// 对子树调用——接口一致，无需特殊处理
	fmt.Println("\n=== 仅查看 assets ===")
	assets.Print("")

	// 对单个文件调用——和目录接口完全相同，这就是 Composite 的价值
	fmt.Println("\n=== 单文件 ===")
	var node Node = &File{name: "go.mod", size: 2}
	node.Print("")
	fmt.Printf("大小: %d KB\n", node.Size())
}
