package main

import (
	"fmt"
	"math/rand"
	"strings"
)

// 场景：游戏地图渲染大量树木
//
// 痛点对比：
//   不用 Flyweight → 每棵 TreeFull 都持有 name/color/texture 副本
//                    100 万棵树 × 1 KB 纹理 ≈ 1 GB，内存直接爆炸
//
//   用 Flyweight  → TreeType（内蕴状态，不可变）按树种仅存一份
//                    100 万棵树共享 3 个 TreeType，每棵只存坐标 + 指针
//                    纹理内存从 GB 级降到 KB 级

// ── 不用 Flyweight 时的结构（用于内存对比）────────────────────────────────────

type TreeFull struct {
	x, y    int
	name    string
	color   string
	texture string // 每个实例独立持有一份纹理数据
}

// ── Flyweight：树种（内蕴状态，不可变，可安全共享）──────────────────────────

type TreeType struct {
	name    string
	color   string
	texture string // 全局只有一份，所有同种树共享
}

func (t *TreeType) Draw(x, y int) {
	fmt.Printf("  [%-4s] 色=%-6s 坐标=(%3d,%3d)\n", t.name, t.color, x, y)
}

// ── Flyweight Factory ─────────────────────────────────────────────────────────
// 保证同种 TreeType 全局只有一份实例，相同 name 直接复用缓存

type TreeTypePool struct {
	pool map[string]*TreeType
}

func NewTreeTypePool() *TreeTypePool {
	return &TreeTypePool{pool: make(map[string]*TreeType)}
}

func (p *TreeTypePool) Get(name, color, texture string) *TreeType {
	if t, ok := p.pool[name]; ok {
		return t // 命中缓存，直接复用
	}
	t := &TreeType{name: name, color: color, texture: texture}
	p.pool[name] = t
	fmt.Printf("  [Pool] 新建 TreeType %q（池大小 %d）\n", name, len(p.pool))
	return t
}

// ── Context：单棵树（外蕴状态，每个实例唯一）────────────────────────────────

type Tree struct {
	x, y     int
	treeType *TreeType // 指向共享 Flyweight，不持有纹理数据副本
}

// ── Forest ────────────────────────────────────────────────────────────────────

type Forest struct {
	trees []*Tree
	pool  *TreeTypePool
}

func NewForest() *Forest {
	return &Forest{pool: NewTreeTypePool()}
}

func (f *Forest) Plant(x, y int, name, color, texture string) {
	tt := f.pool.Get(name, color, texture)
	f.trees = append(f.trees, &Tree{x: x, y: y, treeType: tt})
}

func (f *Forest) Render() {
	for _, t := range f.trees {
		t.treeType.Draw(t.x, t.y)
	}
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	// 3 种树型（texture 模拟纹理数据，实际项目可能是 100 KB 的位图）
	types := []struct{ name, color, texture string }{
		{"橡树", "深绿", strings.Repeat("o", 1024)},
		{"松树", "墨绿", strings.Repeat("p", 1024)},
		{"白桦", "淡绿", strings.Repeat("b", 1024)},
	}
	const textureBytes = 1024 // 每份纹理字节数（模拟 1 KB）

	forest := NewForest()
	rng := rand.New(rand.NewSource(42))
	n := 12 // 植入 12 棵树，但树种只有 3 种

	fmt.Println("━━ 植树（无论种多少棵，Pool 只创建 3 个 TreeType）━━")
	for i := 0; i < n; i++ {
		td := types[rng.Intn(len(types))]
		forest.Plant(rng.Intn(100), rng.Intn(100), td.name, td.color, td.texture)
	}

	fmt.Println()
	fmt.Println("━━ 渲染 ━━")
	forest.Render()

	fmt.Println()
	fmt.Println("━━ 内存对比 ━━")
	typeCount := len(forest.pool.pool)
	treeCount := len(forest.trees)

	withoutFW := treeCount * textureBytes
	withFW := typeCount * textureBytes
	fmt.Printf("  不用 Flyweight：%d 棵 × %d B/纹理 = %d B\n", treeCount, textureBytes, withoutFW)
	fmt.Printf("  用   Flyweight：%d 种 × %d B/纹理 = %d B\n", typeCount, textureBytes, withFW)
	fmt.Printf("  节省 %.1f%%（规模越大越显著：100 万棵时纹理内存从 1 GB → 3 KB）\n",
		float64(withoutFW-withFW)/float64(withoutFW)*100)

	fmt.Println()
	fmt.Println("━━ 验证共享（同种树指向同一 TreeType 实例）━━")
	var oakTrees []*Tree
	for _, t := range forest.trees {
		if t.treeType.name == "橡树" {
			oakTrees = append(oakTrees, t)
		}
	}
	if len(oakTrees) >= 2 {
		fmt.Printf("  橡树 A treeType ptr: %p\n", oakTrees[0].treeType)
		fmt.Printf("  橡树 B treeType ptr: %p\n", oakTrees[1].treeType)
		fmt.Printf("  是同一对象: %v\n", oakTrees[0].treeType == oakTrees[1].treeType)
	}
}
