// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Auth Plugin Example
//
// 示例：自定义认证插件
// 演示如何实现 OnAuth 钩子，对接外部用户系统
//
// 构建：go run ../../tools/build/main.go -dir . -output ./auth_plugin.so
// 测试：go run ../../runner/main.go -plugin ./auth_plugin.so

package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/AXMQ-NET/axmq-plugin-sdk/pluginapi"
)

// AuthPlugin 认证插件实现
type AuthPlugin struct {
	pluginapi.BasePlugin // 嵌入 BasePlugin 获得默认实现

	// 配置
	users map[string]string // username -> password
}

// 确保实现了 Plugin 接口
var _ pluginapi.Plugin = (*AuthPlugin)(nil)

// NewPlugin 插件工厂函数（必须导出）
func NewPlugin() pluginapi.Plugin {
	return &AuthPlugin{
		users: make(map[string]string),
	}
}

// Info 返回插件元信息
func (p *AuthPlugin) Info() pluginapi.PluginMeta {
	return pluginapi.PluginMeta{
		Name:       "auth_plugin",
		Version:    "1.0.0",
		SDKVersion: pluginapi.SDKVersion,
		GoVersion:  runtime.Version(),
		BuildTime:  time.Now().Format(time.RFC3339),
		// HookTimeout: 500 * time.Millisecond, // 如需数据库查询，可设置更长超时
	}
}

// Init 初始化插件
func (p *AuthPlugin) Init(config []byte) error {
	// 默认用户（实际应用中应从配置或外部系统加载）
	p.users["admin"] = "secret"
	p.users["guest"] = "guest123"

	// 如果有配置，解析配置
	if len(config) > 0 {
		var cfg struct {
			Users map[string]string `json:"users"`
		}
		if err := json.Unmarshal(config, &cfg); err == nil && len(cfg.Users) > 0 {
			p.users = cfg.Users
		}
	}

	fmt.Printf("[auth_plugin] Initialized with %d users\n", len(p.users))
	return nil
}

// OnAuth 认证钩子
func (p *AuthPlugin) OnAuth(ctx *pluginapi.AuthContext) (bool, error) {
	// 示例：简单的用户名密码验证
	expectedPass, exists := p.users[ctx.Username]
	if !exists {
		fmt.Printf("[auth_plugin] User not found: %s\n", ctx.Username)
		ctx.ThreatScore = 30 // 用户不存在，记录威胁分
		return false, nil
	}

	if string(ctx.Password) != expectedPass {
		fmt.Printf("[auth_plugin] Wrong password for user: %s\n", ctx.Username)
		ctx.ThreatScore = 50 // 密码错误，较高威胁分（累计达阈值自动拉黑）
		return false, nil
	}

	fmt.Printf("[auth_plugin] Auth success: user=%s, client=%s, ip=%s\n",
		ctx.Username, ctx.ClientID, ctx.IP)
	return true, nil
}

// OnSubscribe 订阅钩子
func (p *AuthPlugin) OnSubscribe(ctx *pluginapi.SubscribeContext) (bool, error) {
	// 示例：禁止订阅 $SYS 主题（除非是 admin）
	if strings.HasPrefix(ctx.Topic, "$SYS") && ctx.Username != "admin" {
		fmt.Printf("[auth_plugin] Denied $SYS subscription for non-admin user: %s\n", ctx.Username)
		return false, nil
	}
	return true, nil
}

// OnPublish 发布钩子（异步通知）
func (p *AuthPlugin) OnPublish(ctx *pluginapi.PublishContext) {
	// 示例：记录所有发布消息（用于审计）
	fmt.Printf("[auth_plugin] Publish: user=%s, topic=%s, qos=%d, size=%d\n",
		ctx.Username, ctx.Topic, ctx.QoS, len(ctx.Payload))
}

// OnDisconnect 断开钩子
func (p *AuthPlugin) OnDisconnect(ctx *pluginapi.DisconnectContext) {
	fmt.Printf("[auth_plugin] Disconnect: user=%s, client=%s, reason=%s\n",
		ctx.Username, ctx.ClientID, ctx.Reason)
}

// Close 关闭插件
func (p *AuthPlugin) Close() error {
	fmt.Println("[auth_plugin] Closing...")
	return nil
}
