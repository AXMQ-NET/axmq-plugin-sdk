// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Logger Plugin Example
//
// 示例：消息日志插件
// 演示如何实现 OnPublish 钩子，记录所有消息到外部系统
//
// 构建：go run ../../tools/build/main.go -dir . -output ./logger_plugin.so
// 测试：go run ../../runner/main.go -plugin ./logger_plugin.so

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/AXMQ-NET/axmq-plugin-sdk/pluginapi"
)

// LoggerPlugin 日志插件实现
type LoggerPlugin struct {
	pluginapi.BasePlugin

	// 配置
	logPath   string
	logFile   *os.File
	mu        sync.Mutex
	msgCount  int64
	connCount int64
}

var _ pluginapi.Plugin = (*LoggerPlugin)(nil)

// NewPlugin 插件工厂函数
func NewPlugin() pluginapi.Plugin {
	return &LoggerPlugin{}
}

// Info 返回插件元信息
func (p *LoggerPlugin) Info() pluginapi.PluginMeta {
	return pluginapi.PluginMeta{
		Name:       "logger_plugin",
		Version:    "1.0.0",
		SDKVersion: pluginapi.SDKVersion,
		GoVersion:  runtime.Version(),
		BuildTime:  time.Now().Format(time.RFC3339),
	}
}

// Init 初始化插件
func (p *LoggerPlugin) Init(config []byte) error {
	// 默认日志路径
	p.logPath = "/tmp/axmq_messages.log"

	// 解析配置
	if len(config) > 0 {
		var cfg struct {
			LogPath string `json:"log_path"`
		}
		if err := json.Unmarshal(config, &cfg); err == nil && cfg.LogPath != "" {
			p.logPath = cfg.LogPath
		}
	}

	// 打开日志文件
	var err error
	p.logFile, err = os.OpenFile(p.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	fmt.Printf("[logger_plugin] Logging to: %s\n", p.logPath)
	return nil
}

// OnAuth 认证钩子 - 记录连接
func (p *LoggerPlugin) OnAuth(ctx *pluginapi.AuthContext) (bool, error) {
	p.mu.Lock()
	p.connCount++
	p.mu.Unlock()

	p.writeLog("CONNECT", map[string]interface{}{
		"client_id": ctx.ClientID,
		"username":  ctx.Username,
		"ip":        ctx.IP,
	})

	return true, nil // 允许所有连接（仅记录）
}

// OnPublish 发布钩子 - 记录消息
func (p *LoggerPlugin) OnPublish(ctx *pluginapi.PublishContext) {
	p.mu.Lock()
	p.msgCount++
	count := p.msgCount
	p.mu.Unlock()

	p.writeLog("PUBLISH", map[string]interface{}{
		"seq":       count,
		"client_id": ctx.ClientID,
		"username":  ctx.Username,
		"topic":     ctx.Topic,
		"qos":       ctx.QoS,
		"retain":    ctx.Retain,
		"size":      len(ctx.Payload),
	})
}

// OnDisconnect 断开钩子 - 记录断开
func (p *LoggerPlugin) OnDisconnect(ctx *pluginapi.DisconnectContext) {
	p.writeLog("DISCONNECT", map[string]interface{}{
		"client_id": ctx.ClientID,
		"username":  ctx.Username,
		"reason":    ctx.Reason,
	})
}

// writeLog 写入日志
func (p *LoggerPlugin) writeLog(event string, data map[string]interface{}) {
	if p.logFile == nil {
		return
	}

	entry := map[string]interface{}{
		"time":  time.Now().Format(time.RFC3339Nano),
		"event": event,
	}
	for k, v := range data {
		entry[k] = v
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.logFile.Write(append(line, '\n'))
}

// Close 关闭插件
func (p *LoggerPlugin) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Printf("[logger_plugin] Closing. Total messages: %d, connections: %d\n",
		p.msgCount, p.connCount)

	if p.logFile != nil {
		p.logFile.Close()
		p.logFile = nil
	}
	return nil
}
