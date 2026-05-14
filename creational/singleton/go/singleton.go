package main

import (
	"fmt"
	"sync"
	"time"
)

// 场景：应用全局配置中心
//
// ──────────────────────────────────────────────────────────────────────────────
// 痛点对比
// ──────────────────────────────────────────────────────────────────────────────
//
// ✗ 不用单例（每次直接 new Config{}）
//
//	func GetConfig() *Config {
//	    cfg := &Config{}          // 每次都 new，地址各不同
//	    loadFromDisk(cfg)         // 痛点①：耗时 I/O 重复执行，启动慢/运行慢
//	    return cfg
//	}
//
//	// 痛点②：各模块拿到不同实例，修改互不可见
//	GetConfig().LogLevel = "debug"    // 只改了这一份，其他模块毫不知情
//
//	// 痛点③：多 goroutine 并发调用时各自 new，可能拿到状态不一致的副本
//	//        若初始化包含写全局变量，还会有数据竞争（race condition）
//
// ✓ 用单例（sync.Once 惰性初始化）
//
//	// 痛点① → 只初始化一次，节省 I/O
//	// 痛点② → 所有调用方共享同一指针，改一处全局生效
//	// 痛点③ → sync.Once 内置互斥，并发安全无需额外加锁
//
// ──────────────────────────────────────────────────────────────────────────────
//
// Singleton 的三个核心价值
//   1. 创建成本高（读文件/远程拉取），只应发生一次
//   2. 持有全局共享状态，改一处、全局生效
//   3. 并发安全，任何 goroutine 拿到的是同一个实例

type Config struct {
	DSN         string
	MaxConn     int
	LogLevel    string
	ServiceName string
}

var (
	once   sync.Once
	config *Config
)

func GetConfig() *Config {
	// 对比痛点①：若改成 return &Config{...}，每次调用都执行下方耗时 I/O
	once.Do(func() {
		// 模拟从配置文件/远程配置中心加载（耗时操作）
		fmt.Println("  [init] 正在从磁盘加载配置...")
		time.Sleep(20 * time.Millisecond)
		config = &Config{
			DSN:         "postgres://localhost:5432/mydb",
			MaxConn:     20,
			LogLevel:    "info",
			ServiceName: "order-service",
		}
		fmt.Printf("  [init] 加载完成，实例地址 %p\n\n", config)
	})
	return config
}

// ── 模拟三个独立模块，各自在需要时按需取配置 ─────────────────────────────────

func dbModule() string {
	cfg := GetConfig()
	return fmt.Sprintf("DB  模块 → DSN=%s  MaxConn=%d  (addr=%p)", cfg.DSN, cfg.MaxConn, cfg)
}

func logModule() string {
	cfg := GetConfig()
	return fmt.Sprintf("Log 模块 → LogLevel=%s  (addr=%p)", cfg.LogLevel, cfg)
}

func apiModule() string {
	cfg := GetConfig()
	return fmt.Sprintf("API 模块 → ServiceName=%s  (addr=%p)", cfg.ServiceName, cfg)
}

func main() {
	// ── 价值 1 & 3：5 个 goroutine 并发初始化，"加载配置"只打印一次 ──────────
	fmt.Println("━━ 并发安全：sync.Once 保证只初始化一次 ━━")
	var wg sync.WaitGroup
	for i := range 5 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cfg := GetConfig()
			fmt.Printf("  goroutine %d 拿到配置 %p\n", n, cfg)
		}(i)
	}
	wg.Wait()

	// ── 价值 2：各模块拿到同一份实例，地址完全一致 ──────────────────────────
	fmt.Println("━━ 全局唯一：各模块共享同一实例 ━━")
	fmt.Println(" ", dbModule())
	fmt.Println(" ", logModule())
	fmt.Println(" ", apiModule())

	// ── 价值 2（续）：运行时修改，无需重启，全局立即生效 ──────────────────────
	// 对比痛点②：若每次 GetConfig() 都 new 一个，下方修改仅影响当前实例，
	//             其他模块调用 logModule() 时仍拿到旧值 "info"，状态撕裂
	fmt.Println("\n━━ 共享状态：改一处，全局生效 ━━")
	fmt.Println("  修改前：", logModule())
	GetConfig().LogLevel = "debug" // 直接改单例字段
	fmt.Println("  修改后：", logModule())
	fmt.Println("  API模块也能看到同一实例：", fmt.Sprintf("%p", GetConfig()))
}
