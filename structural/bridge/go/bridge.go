package main

import "fmt"

// 场景：绘图应用支持多种形状 × 多种渲染方式（向量 / 像素）
//
// 痛点对比：
//   不用桥接 → 每种组合单独建类：VectorCircle、PixelCircle、VectorRect、PixelRect…
//               M 种形状 × N 种渲染 = M×N 个类；新增一种渲染方式要新增 M 个类
//
//   用桥接  → 形状（抽象层）持有 Renderer（实现层）引用，两个维度独立扩展
//               M + N 个类搞定；新增渲染方式只加 1 个 Renderer，形状代码零改动

// Renderer 实现层接口：定义底层渲染能力
type Renderer interface {
	RenderCircle(radius float64)
	RenderRect(width, height float64)
}

// --- 实现层 ---

type VectorRenderer struct{}

func (v *VectorRenderer) RenderCircle(radius float64) {
	fmt.Printf("向量绘制圆形，半径 %.1f\n", radius)
}
func (v *VectorRenderer) RenderRect(width, height float64) {
	fmt.Printf("向量绘制矩形，%.1f × %.1f\n", width, height)
}

type PixelRenderer struct{}

func (p *PixelRenderer) RenderCircle(radius float64) {
	fmt.Printf("像素绘制圆形，半径 %.1f\n", radius)
}
func (p *PixelRenderer) RenderRect(width, height float64) {
	fmt.Printf("像素绘制矩形，%.1f × %.1f\n", width, height)
}

// --- 抽象层 ---

// Shape 基础抽象，持有实现层引用
type Shape struct {
	renderer Renderer
}

// Circle 精化抽象：圆形
type Circle struct {
	Shape
	radius float64
}

func NewCircle(r Renderer, radius float64) *Circle {
	return &Circle{Shape: Shape{renderer: r}, radius: radius}
}

func (c *Circle) Draw() {
	c.renderer.RenderCircle(c.radius)
}

// Rect 精化抽象：矩形
type Rect struct {
	Shape
	width, height float64
}

func NewRect(r Renderer, width, height float64) *Rect {
	return &Rect{Shape: Shape{renderer: r}, width: width, height: height}
}

func (r *Rect) Draw() {
	r.renderer.RenderRect(r.width, r.height)
}

func main() {
	vector := &VectorRenderer{}
	pixel := &PixelRenderer{}

	// 形状 × 渲染器，自由组合，互不影响
	NewCircle(vector, 5).Draw()
	NewCircle(pixel, 5).Draw()
	NewRect(vector, 4, 3).Draw()
	NewRect(pixel, 4, 3).Draw()
}
