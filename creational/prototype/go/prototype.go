package main

import (
	"fmt"
	"time"
)

// 问题背景：游戏关卡需要批量生成同类敌人。
// 每个敌人首次初始化须从"配置库"加载技能、装备等数据（耗时操作）。
//
// 原型模式的解法：
//   1. 预先初始化一个完整的原型对象（只付一次昂贵的初始化代价）。
//   2. 后续通过 Clone() 快速复制，跳过重复的耗时初始化。
//   3. Clone() 必须做深拷贝，确保各克隆体修改自身数据时不会污染原型。

type Enemy struct {
	Name      string
	HP        int
	Skills    []string
	Equipment map[string]int // 装备名 -> 强化等级
}

// Clone 深拷贝，切片和 map 独立分配，保证克隆体与原型互不影响
func (e *Enemy) Clone() *Enemy {
	skills := make([]string, len(e.Skills))
	copy(skills, e.Skills)

	equipment := make(map[string]int, len(e.Equipment))
	for k, v := range e.Equipment {
		equipment[k] = v
	}

	return &Enemy{
		Name:      e.Name,
		HP:        e.HP,
		Skills:    skills,
		Equipment: equipment,
	}
}

// loadFromConfig 模拟从配置库加载敌人数据的耗时初始化
func loadFromConfig(name string) *Enemy {
	time.Sleep(100 * time.Millisecond)
	return &Enemy{
		Name:      name,
		HP:        1000,
		Skills:    []string{"冰冻", "火球", "雷击"},
		Equipment: map[string]int{"长剑": 5, "盔甲": 3},
	}
}

func main() {
	const spawnCount = 5

	// ── 不用原型：每次都重新加载配置（耗时随数量线性增长）──
	fmt.Printf("=== 不用原型模式：每次重新初始化（共 %d 次）===\n", spawnCount)
	start := time.Now()
	for i := 0; i < spawnCount; i++ {
		loadFromConfig("兽人")
	}
	fmt.Printf("耗时：%v\n\n", time.Since(start))

	// ── 用原型：只初始化一次，克隆复用 ──
	fmt.Printf("=== 原型模式：初始化一次，克隆 %d 份 ===\n", spawnCount)
	start = time.Now()
	prototype := loadFromConfig("兽人") // 仅一次耗时加载
	enemies := make([]*Enemy, spawnCount)
	for i := 0; i < spawnCount; i++ {
		enemies[i] = prototype.Clone()
	}
	fmt.Printf("耗时：%v\n\n", time.Since(start))

	// ── 深拷贝验证：修改克隆体不影响原型 ──
	fmt.Println("=== 深拷贝：克隆体可以独立修改，原型不受影响 ===")
	elite := prototype.Clone()
	elite.Name = "精英兽人"
	elite.Skills = append(elite.Skills, "狂暴") // 为精英兽人添加专属技能
	elite.Equipment["长剑"] = 10                // 升级精英兽人的武器

	fmt.Printf("原型  — 名称: %-8s 技能: %v 长剑等级: %d\n",
		prototype.Name, prototype.Skills, prototype.Equipment["长剑"])
	fmt.Printf("克隆体— 名称: %-8s 技能: %v 长剑等级: %d\n",
		elite.Name, elite.Skills, elite.Equipment["长剑"])
}
