// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Plugin Metadata

package pluginapi

import "time"

// 默认超时配置
const (
	DefaultHookTimeout = 100 * time.Millisecond // 默认钩子超时
	MaxHookTimeout     = 30 * time.Second       // 最大允许超时
)

// PluginMeta 插件元信息
// 用于加载时校验插件与主程序的兼容性
type PluginMeta struct {
	Name        string        `json:"name"`                   // 插件名称
	Version     string        `json:"version"`                // 插件版本
	SDKVersion  string        `json:"sdk_version"`            // SDK 版本（必须与主程序一致）
	GoVersion   string        `json:"go_version"`             // 编译时的 Go 版本
	BuildTime   string        `json:"build_time"`             // 构建时间 (RFC3339)
	HookTimeout time.Duration `json:"hook_timeout,omitempty"` // 钩子超时时间（0 表示使用默认 100ms）
}

// GetHookTimeout 获取有效的超时时间
func (m *PluginMeta) GetHookTimeout() time.Duration {
	if m.HookTimeout <= 0 {
		return DefaultHookTimeout
	}
	if m.HookTimeout > MaxHookTimeout {
		return MaxHookTimeout
	}
	return m.HookTimeout
}

// Validate 校验元信息
func (m *PluginMeta) Validate() error {
	if m.Name == "" {
		return ErrInvalidPluginName
	}
	if m.SDKVersion == "" {
		return ErrMissingSDKVersion
	}
	if m.SDKVersion != SDKVersion {
		return ErrSDKVersionMismatch
	}
	return nil
}
