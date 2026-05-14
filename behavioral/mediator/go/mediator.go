package main

import "fmt"

// 场景：聊天室消息调度
//
// 痛点对比：
//   不用中介者 → 每个用户持有房间内所有其他用户的引用
//     Alice 想广播，需要遍历 [Bob, Charlie, ...] 逐一调用 Receive()
//     新用户加入时，所有现有用户都要更新引用列表，N 人产生 N×(N-1) 条耦合
//     禁言、过滤、私聊等逻辑只能散落在每个用户对象里，无法统一管控
//
//   用中介者   → 每个用户只认识 ChatRoom，完全不感知彼此的存在
//     发消息只调 room.Send()，由 ChatRoom 统一负责转发、过滤、广播
//     加人 / 踢人 / 禁言只修改 ChatRoom，所有用户对象零改动
//     N×N 耦合缩减为 N×1，新增规则（如敏感词过滤）只在中介者加一处逻辑

// ── 中介者接口 ────────────────────────────────────────────────────────────────

type Mediator interface {
	Join(user *User)
	Leave(user *User)
	// to="" 表示广播，否则为私聊目标用户名
	Send(from *User, to string, msg string)
}

// ── 同事类（Colleague）────────────────────────────────────────────────────────

type User struct {
	Name string
	room Mediator // 只持有中介者，不直接引用任何其他用户
}

func NewUser(name string) *User { return &User{Name: name} }

func (u *User) JoinRoom(room Mediator) {
	u.room = room
	room.Join(u)
}

func (u *User) LeaveRoom() {
	u.room.Leave(u)
	u.room = nil
}

func (u *User) Broadcast(msg string) {
	u.room.Send(u, "", msg)
}

func (u *User) PrivateTo(target, msg string) {
	u.room.Send(u, target, msg)
}

// Receive 由中介者回调，用户被动接收消息，无需关心来源路由
func (u *User) Receive(from, msg string, private bool) {
	label := ""
	if private {
		label = "[私聊] "
	}
	fmt.Printf("    %-10s 收到 %s%s：%s\n", u.Name, label, from, msg)
}

// ── 具体中介者：ChatRoom ──────────────────────────────────────────────────────

type ChatRoom struct {
	name    string
	users   map[string]*User
	muted   map[string]bool
	blocked []string // 敏感词列表
}

func NewChatRoom(name string, blocked []string) *ChatRoom {
	return &ChatRoom{
		name:    name,
		users:   make(map[string]*User),
		muted:   make(map[string]bool),
		blocked: blocked,
	}
}

func (r *ChatRoom) Join(user *User) {
	r.users[user.Name] = user
	fmt.Printf("[%s] ✦ %s 加入（当前 %d 人）\n", r.name, user.Name, len(r.users))
}

func (r *ChatRoom) Leave(user *User) {
	delete(r.users, user.Name)
	fmt.Printf("[%s] ✦ %s 离开（当前 %d 人）\n", r.name, user.Name, len(r.users))
}

func (r *ChatRoom) Mute(name string) {
	r.muted[name] = true
	fmt.Printf("[%s] ⚑ %s 被禁言\n", r.name, name)
}

func (r *ChatRoom) Send(from *User, to, msg string) {
	if r.muted[from.Name] {
		fmt.Printf("[%s] ✗ 拦截（禁言）：%s → %q\n", r.name, from.Name, msg)
		return
	}
	for _, word := range r.blocked {
		if contains(msg, word) {
			fmt.Printf("[%s] ✗ 拦截（敏感词 %q）：%s → %q\n", r.name, word, from.Name, msg)
			return
		}
	}

	if to == "" {
		// 广播：中介者统一遍历，用户对象完全不参与路由
		fmt.Printf("[%s] ↦ %s 广播：%s\n", r.name, from.Name, msg)
		for name, u := range r.users {
			if name != from.Name {
				u.Receive(from.Name, msg, false)
			}
		}
		return
	}

	// 私聊
	target, ok := r.users[to]
	if !ok {
		fmt.Printf("[%s] ✗ 私聊失败：%s 不在房间\n", r.name, to)
		return
	}
	fmt.Printf("[%s] ↦ %s → %s（私聊）\n", r.name, from.Name, to)
	target.Receive(from.Name, msg, true)
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// ── main ─────────────────────────────────────────────────────────────────────

func main() {
	room := NewChatRoom("Go 技术交流群", []string{"广告"})

	alice := NewUser("Alice")
	bob := NewUser("Bob")
	charlie := NewUser("Charlie")

	alice.JoinRoom(room)
	bob.JoinRoom(room)
	charlie.JoinRoom(room)

	fmt.Println("\n━━ 广播：所有人收到，中介者统一路由 ━━")
	alice.Broadcast("大家好！有人研究过 Mediator 模式吗？")

	fmt.Println("\n━━ 私聊：只有目标用户收到 ━━")
	bob.PrivateTo("Alice", "我研究过，我们可以深聊")
	charlie.PrivateTo("Dave", "Dave 你在吗？") // Dave 不在房间，中介者报错

	fmt.Println("\n━━ 敏感词过滤：中介者统一拦截，无需每个用户自己判断 ━━")
	charlie.Broadcast("点击链接领取广告福利")

	fmt.Println("\n━━ 禁言：只改中介者状态，用户对象不感知 ━━")
	room.Mute("Charlie")
	charlie.Broadcast("我还可以说话！") // 被中介者拦截

	fmt.Println("\n━━ 用户离开：只改中介者，其他用户无需更新引用 ━━")
	bob.LeaveRoom()
	alice.Broadcast("Bob 走了，还有人吗？") // Charlie 禁言中，只能被动接收
}
