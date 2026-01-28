// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Hook Contexts

package pluginapi

// AuthContext 认证上下文
// 在 CONNECT 报文处理时传递给 OnAuth 钩子
type AuthContext struct {
	// 输入字段（主程序填充）
	ClientID string // 客户端 ID
	Username string // 用户名
	Password []byte // 密码
	IP       string // 客户端 IP 地址

	// 输出字段（插件可设置）
	ThreatScore int // 威胁计分（0=正常，>0=可疑，累计达阈值自动拉黑）
}

// SubscribeContext 订阅上下文
// 在 SUBSCRIBE 报文处理时传递给 OnSubscribe 钩子
type SubscribeContext struct {
	// 输入字段（主程序填充）
	ClientID string // 客户端 ID
	Username string // 用户名
	Topic    string // 订阅主题（可能包含通配符）
	QoS      uint8  // 请求的 QoS 等级
	IP       string // 客户端 IP 地址

	// 输出字段（插件可设置）
	ThreatScore int // 威胁计分
}

// PublishContext 发布上下文
// 在 PUBLISH 报文处理时传递给 OnPublish 钩子
// 注意：此钩子为异步调用，不阻塞主流程
type PublishContext struct {
	ClientID string // 发布者客户端 ID
	Username string // 发布者用户名
	Topic    string // 发布主题
	Payload  []byte // 消息内容（只读副本）
	QoS      uint8  // QoS 等级
	Retain   bool   // 是否为保留消息
}

// DisconnectContext 断开上下文
// 在客户端断开连接时传递给 OnDisconnect 钩子
type DisconnectContext struct {
	ClientID string // 客户端 ID
	Username string // 用户名
	Reason   string // 断开原因（graceful/timeout/error）
}
