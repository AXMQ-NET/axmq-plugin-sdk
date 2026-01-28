// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Error Definitions

package pluginapi

import "errors"

var (
	// 元数据校验错误
	ErrInvalidPluginName  = errors.New("plugin name is empty")
	ErrMissingSDKVersion  = errors.New("sdk version is missing")
	ErrSDKVersionMismatch = errors.New("sdk version mismatch between plugin and host")
	ErrGoVersionMismatch  = errors.New("go version mismatch between plugin and host")

	// 加载错误
	ErrPluginNotFound     = errors.New("plugin file not found")
	ErrMetaNotFound       = errors.New("plugin meta file not found")
	ErrSymbolNotFound     = errors.New("plugin does not export 'NewPlugin' symbol")
	ErrInvalidPluginType  = errors.New("plugin symbol is not of type func() Plugin")
	ErrPluginInitFailed   = errors.New("plugin initialization failed")
	ErrPluginAlreadyExist = errors.New("plugin with same name already loaded")
)
