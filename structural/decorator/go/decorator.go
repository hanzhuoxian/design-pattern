package main

import (
	"fmt"
	"strings"
	"time"
)

// 场景：数据存储服务的能力叠加
//
// 痛点对比：
//   不用 Decorator → 把日志、缓存、压缩全塞进 DBStorage
//                    三个关注点耦合在一个类里，无法单独开关，
//                    也无法把"日志"或"缓存"复用到 FileStorage / S3Storage
//
//   用 Decorator  → 每层只做一件事，按需组合：
//                    cache(log(compress(db)))  or  log(db)  or  cache(file)
//                    调用方面对的始终是同一个 Storage 接口

// ── Component 接口 ────────────────────────────────────────────────────────────

type Storage interface {
	Save(key, value string) error
	Load(key string) (string, error)
}

// ── Concrete Component：真正的 DB 存储 ────────────────────────────────────────

type DBStorage struct {
	data map[string]string
}

func NewDBStorage() *DBStorage {
	return &DBStorage{data: make(map[string]string)}
}

func (s *DBStorage) Save(key, value string) error {
	s.data[key] = value
	return nil
}

func (s *DBStorage) Load(key string) (string, error) {
	v, ok := s.data[key]
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}
	return v, nil
}

// ── Decorator 1：日志 ─────────────────────────────────────────────────────────

type LoggingStorage struct {
	inner Storage
}

func WithLogging(s Storage) Storage {
	return &LoggingStorage{inner: s}
}

func (s *LoggingStorage) Save(key, value string) error {
	start := time.Now()
	err := s.inner.Save(key, value)
	fmt.Printf("[LOG]   Save  key=%-12q  elapsed=%v  err=%v\n", key, time.Since(start), err)
	return err
}

func (s *LoggingStorage) Load(key string) (string, error) {
	start := time.Now()
	v, err := s.inner.Load(key)
	fmt.Printf("[LOG]   Load  key=%-12q  elapsed=%v  err=%v\n", key, time.Since(start), err)
	return v, err
}

// ── Decorator 2：内存缓存 ─────────────────────────────────────────────────────

type CachingStorage struct {
	inner Storage
	cache map[string]string
}

func WithCaching(s Storage) Storage {
	return &CachingStorage{inner: s, cache: make(map[string]string)}
}

// 写穿：Save 时同步使缓存失效，保证一致性
func (s *CachingStorage) Save(key, value string) error {
	delete(s.cache, key)
	return s.inner.Save(key, value)
}

func (s *CachingStorage) Load(key string) (string, error) {
	if v, ok := s.cache[key]; ok {
		fmt.Printf("[CACHE] hit   key=%-12q\n", key)
		return v, nil
	}
	v, err := s.inner.Load(key)
	if err == nil {
		s.cache[key] = v
	}
	return v, err
}

// ── Decorator 3：压缩（用 Base64 风格标记模拟，聚焦结构而非算法细节）─────────

type CompressStorage struct {
	inner Storage
}

func WithCompression(s Storage) Storage {
	return &CompressStorage{inner: s}
}

func compress(v string) string {
	// 模拟压缩：实际可替换为 gzip.NewWriter
	return "[gz:" + strings.ReplaceAll(v, " ", "_") + "]"
}

func decompress(v string) string {
	v = strings.TrimPrefix(v, "[gz:")
	v = strings.TrimSuffix(v, "]")
	return strings.ReplaceAll(v, "_", " ")
}

func (s *CompressStorage) Save(key, value string) error {
	return s.inner.Save(key, compress(value))
}

func (s *CompressStorage) Load(key string) (string, error) {
	v, err := s.inner.Load(key)
	if err != nil {
		return "", err
	}
	return decompress(v), nil
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	// 1. 裸存储：没有任何附加能力
	fmt.Println("━━ 裸 DBStorage ━━")
	plain := NewDBStorage()
	_ = plain.Save("user:1", "Alice")
	v, _ := plain.Load("user:1")
	fmt.Println("Loaded:", v)
	fmt.Println()

	// 2. 只加日志：单一关注点，不影响其他任何逻辑
	fmt.Println("━━ +日志 ━━")
	logged := WithLogging(NewDBStorage())
	_ = logged.Save("user:1", "Alice")
	v, _ = logged.Load("user:1")
	fmt.Println("Loaded:", v)
	fmt.Println()

	// 3. 日志 + 缓存叠加：顺序决定行为
	//    调用链：CachingStorage → (miss) → LoggingStorage → DBStorage
	//    第二次 Load 在 CachingStorage 命中缓存，日志层根本不会被触碰
	fmt.Println("━━ +日志 +缓存（第二次 Load 走缓存，日志沉默）━━")
	cached := WithCaching(WithLogging(NewDBStorage()))
	_ = cached.Save("user:1", "Alice")
	fmt.Print("第1次 Load: ")
	v, _ = cached.Load("user:1")
	fmt.Println("Loaded:", v)
	fmt.Print("第2次 Load: ")
	v, _ = cached.Load("user:1")
	fmt.Println("Loaded:", v)
	fmt.Println()

	// 4. 完整组合：压缩 + 日志 + 缓存
	//    调用链：CachingStorage → LoggingStorage → CompressStorage → DBStorage
	//    DB 里存压缩后的内容，用户拿到的是解压后的原始数据，中间层透明
	fmt.Println("━━ +压缩 +日志 +缓存（生产级叠加，每层职责不重叠）━━")
	full := WithCaching(WithLogging(WithCompression(NewDBStorage())))
	_ = full.Save("order:99", `{"item": "book", "qty": 2}`)
	v, _ = full.Load("order:99")
	fmt.Println("Loaded:", v)
	fmt.Println()

	// 5. 同一套装饰器可以包装不同 Storage 实现
	//    体现 Decorator 的复用价值：日志/缓存逻辑零改动，换掉底层存储即可
	fmt.Println("━━ 复用同一套装饰器包装 FileStorage（接口统一）━━")
	type FileStorage struct{ DBStorage } // 用 DBStorage 模拟，聚焦结构
	file := &FileStorage{DBStorage: *NewDBStorage()}
	loggedFile := WithLogging(file)
	_ = loggedFile.Save("config:app", "debug=true")
	v, _ = loggedFile.Load("config:app")
	fmt.Println("Loaded:", v)
}
