// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Plugin Interface Definition

package pluginapi

// Plugin 插件必须实现的接口
// 每个插件必须导出一个 NewPlugin 函数，签名为：func NewPlugin() Plugin
type Plugin interface {
	// Info 返回插件元信息
	// 加载时调用，用于校验插件与主程序的兼容性
	Info() PluginMeta

	// Init 插件初始化
	// 加载成功后调用一次，config 为插件配置内容（可为空）
	// 返回 error 将导致插件加载失败
	Init(config []byte) error

	// Close 插件关闭
	// 卸载或主程序退出前调用，用于清理资源
	Close() error

	// === 低频钩子 ===

	// OnAuth 认证钩子
	// 触发时机：CONNECT 报文处理时，在内置认证之后调用
	// 返回值：
	//   - allow=true:  允许连接
	//   - allow=false: 拒绝连接（将返回 CONNACK 0x05）
	//   - err!=nil:    发生错误，记录日志但不影响连接结果
	OnAuth(ctx *AuthContext) (allow bool, err error)

	// OnSubscribe 订阅钩子
	// 触发时机：SUBSCRIBE 报文处理时，在内置 ACL 检查之后调用
	// 返回值：
	//   - allow=true:  允许订阅
	//   - allow=false: 拒绝订阅（SUBACK 返回 0x80）
	//   - err!=nil:    发生错误，记录日志但不影响订阅结果
	OnSubscribe(ctx *SubscribeContext) (allow bool, err error)

	// OnPublish 发布钩子
	// 触发时机：PUBLISH 报文处理后，异步调用（不阻塞消息分发）
	// 用途：消息审计、日志记录、外部系统通知等
	// 注意：此钩子不能拒绝或修改消息，仅用于通知
	OnPublish(ctx *PublishContext)

	// OnDisconnect 断开钩子
	// 触发时机：客户端断开连接后
	// 用途：清理资源、审计日志、通知外部系统
	OnDisconnect(ctx *DisconnectContext)
}

// BasePlugin 提供默认空实现
// 插件可以嵌入此结构体，只覆盖需要的钩子
type BasePlugin struct{}

func (BasePlugin) OnAuth(ctx *AuthContext) (bool, error)           { return true, nil }
func (BasePlugin) OnSubscribe(ctx *SubscribeContext) (bool, error) { return true, nil }
func (BasePlugin) OnPublish(ctx *PublishContext)                   {}
func (BasePlugin) OnDisconnect(ctx *DisconnectContext)             {}
func (BasePlugin) Close() error                                    { return nil }
