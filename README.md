# AXMQ Plugin SDK

AXMQ Go 原生插件开发工具包。

## 特性

- **性能优先**：Go 原生插件，零 FFI 开销
- **安全加载**：多层版本校验，防止运行时崩溃
- **热加载**：支持插件新增、更新、删除，无需重启服务
- **超时保护**：防止慢插件阻塞系统，可自定义超时
- **防攻击**：威胁计分系统，自动拉黑恶意 IP
- **本地调试**：无需运行完整 AXMQ 即可测试插件

## 快速开始

### 1. 创建插件项目

```bash
mkdir my_plugin && cd my_plugin
go mod init my_plugin
go get github.com/AXMQ-NET/axmq-plugin-sdk@latest
```

### 2. 实现插件

```go
package main

import (
    "runtime"
    "time"
    "github.com/AXMQ-NET/axmq-plugin-sdk/pluginapi"
)

type MyPlugin struct {
    pluginapi.BasePlugin // 嵌入获得默认实现
}

// NewPlugin 必须导出的工厂函数
func NewPlugin() pluginapi.Plugin {
    return &MyPlugin{}
}

func (p *MyPlugin) Info() pluginapi.PluginMeta {
    return pluginapi.PluginMeta{
        Name:        "my_plugin",
        Version:     "1.0.0",
        SDKVersion:  pluginapi.SDKVersion,
        GoVersion:   runtime.Version(),
        BuildTime:   time.Now().Format(time.RFC3339),
        HookTimeout: 200 * time.Millisecond, // 可选：自定义超时（默认 100ms）
    }
}

func (p *MyPlugin) Init(config []byte) error {
    return nil
}

func (p *MyPlugin) OnAuth(ctx *pluginapi.AuthContext) (bool, error) {
    // 认证失败时设置威胁计分，累计达阈值自动拉黑
    if ctx.Username == "" {
        ctx.ThreatScore = 50
        return false, nil
    }
    return true, nil
}
```

### 3. 构建插件

```bash
go build -buildmode=plugin -o my_plugin.so .
```

或使用 SDK 工具（自动生成元数据）：

```bash
go run github.com/AXMQ-NET/axmq-plugin-sdk/tools/build@latest -dir . -output ./my_plugin.so
```

### 4. 本地测试

```bash
go run github.com/AXMQ-NET/axmq-plugin-sdk/runner@latest -plugin ./my_plugin.so
```

交互式命令：
```
> auth client001 admin secret 192.168.1.1
> subscribe client001 admin sensor/+/data 1
> publish client001 admin sensor/1/data hello 0
> disconnect client001 admin
```

### 5. 部署（支持热加载）

```bash
# 复制到 AXMQ 插件目录，自动加载
# AXMQ启动时会提示正真的路径：例如：Watcher started: ./data/plugins
cp my_plugin.so ./data/plugins
cp my_plugin.so.meta.json ./data/plugins

# 可选：插件配置
cp my_plugin.yml ./data/plugins
```

## 钩子说明

| 钩子 | 触发时机 | 阻塞 | 超时策略 | 用途 |
|------|---------|------|----------|------|
| `OnAuth` | CONNECT 认证 | 是 | 超时→拒绝 | 自定义认证、外部系统对接 |
| `OnSubscribe` | SUBSCRIBE 处理 | 是 | 超时→拒绝 | 订阅 ACL、审计 |
| `OnPublish` | PUBLISH 处理后 | 否 | 超时→跳过 | 消息审计、日志、通知（由 `OnPublishAsync` 异步触发） |
| `OnDisconnect` | 连接断开 | 否 | 超时→跳过 | 清理、审计 |

**执行特性**：
- 多插件并行执行，取最慢者耗时
- OnAuth/OnSubscribe 任一插件拒绝即终止
- 每个插件有独立超时，互不影响
- 插件 panic 不影响主程序
- 频繁出错的插件会被自动禁用（熔断保护）
- **OnPublish 异步通知**：主程序使用 `OnPublishAsync` 触发，不阻塞消息分发

### OnPublishAsync 示例

`OnPublish` 默认由主程序异步调用，插件只需实现 `OnPublish`：

```go
func (p *MyPlugin) OnPublish(ctx *pluginapi.PublishContext) {
    // 异步触发的通知钩子，不能阻塞业务主流程
    log.Printf("publish: user=%s topic=%s qos=%d", ctx.Username, ctx.Topic, ctx.QoS)
}
```

## 超时配置

```go
func (p *MyPlugin) Info() pluginapi.PluginMeta {
    return pluginapi.PluginMeta{
        // ...
        HookTimeout: 500 * time.Millisecond, // 默认 100ms，最大 30s
    }
}
```

| 场景 | 建议超时 |
|------|----------|
| 内存查询 | 100ms（默认） |
| 数据库查询 | 200-500ms |
| 外部 HTTP 调用 | 1-5s |

**安全策略**：
- `OnAuth` / `OnSubscribe`：超时**拒绝**请求（防止 DDoS 绕过认证）
- `OnPublish` / `OnDisconnect`：超时**跳过**该插件（不影响业务）

## 威胁计分（防攻击）

插件可设置 `ThreatScore`，主程序累加同一 IP 的分数，达到阈值自动拉黑：

```go
func (p *MyPlugin) OnAuth(ctx *pluginapi.AuthContext) (bool, error) {
    if !userExists(ctx.Username) {
        ctx.ThreatScore = 30  // 用户不存在
        return false, nil
    }
    if !checkPassword(ctx.Username, ctx.Password) {
        ctx.ThreatScore = 50  // 密码错误
        return false, nil
    }
    return true, nil
}
```

| 场景 | 建议分数 | 说明 |
|------|----------|------|
| 用户不存在 | 30 | 可能是探测 |
| 密码错误 | 50 | 暴力破解特征 |
| 已知恶意特征 | 100 | 立即触发封禁 |
| 超时 | 100 | 系统自动设置 |

**注意**：默认阈值 100，累计达到后自动封禁 IP。

## 热加载

AXMQ 监控插件目录，支持运行时管理：

| 操作 | 触发方式 | 行为 |
|------|----------|------|
| 新增 | 复制 `.so` 到目录 | 自动加载启用 |
| 更新 | 覆盖 `.so` 文件 | 优雅切换，零事件丢失 |
| 删除 | 删除 `.so` 文件 | 优雅下线 |

## 熔断机制

为保护系统稳定性，频繁出错的插件会被自动禁用：

| 指标 | 阈值 | 触发动作 |
|------|------|----------|
| Panic 次数 | ≥ 3 | 自动禁用 |
| 错误次数 | ≥ 10 | 自动禁用 |
| 超时 | 计入错误 | 累计触发 |

**恢复方式**：更新 `.so` 文件会热加载新版本并重置计数器。

**建议**：
- 充分测试插件，避免 panic
- 合理设置 `HookTimeout`，避免频繁超时
- 通过 AXMQ 日志监控插件健康状态

## 目录结构

```
axmq-plugin-sdk/
├── pluginapi/          # 核心接口
│   ├── api.go          # Plugin 接口
│   ├── context.go      # 钩子上下文（含 ThreatScore）
│   ├── meta.go         # 元数据（含 HookTimeout）
│   ├── errors.go       # 错误定义
│   └── version.go      # SDK 版本
├── runner/             # 本地调试器
├── tools/build/        # 构建工具
└── examples/
    ├── auth_plugin/    # 认证示例
    └── logger_plugin/  # 日志示例
```

## 版本兼容性

**重要**：插件和 AXMQ 主程序必须使用：
- 相同的 Go 版本
- 相同的 SDK 版本
- 相同的操作系统和架构

SDK 版本校验在加载时自动进行，不兼容的插件将被拒绝。

## License

Apache License 2.0
