# Logger Plugin Example

消息日志插件示例，演示如何实现 OnPublish 钩子记录所有消息。

## 功能

- 记录所有 CONNECT 事件
- 记录所有 PUBLISH 消息（Topic、QoS、大小等）
- 记录所有 DISCONNECT 事件
- 输出 JSON Lines 格式日志

## 构建

```bash
cd examples/logger_plugin
go run ../../tools/build/main.go -dir . -output ./logger_plugin.so
```

## 本地测试

```bash
go run ../../runner/main.go -plugin ./logger_plugin.so
```

## 配置

创建 `config.json`:

```json
{
  "log_path": "/var/log/axmq/messages.log"
}
```

## 日志格式

```json
{"time":"2025-01-28T10:30:00.123456789Z","event":"CONNECT","client_id":"client001","username":"admin","ip":"192.168.1.1"}
{"time":"2025-01-28T10:30:01.234567890Z","event":"PUBLISH","seq":1,"client_id":"client001","username":"admin","topic":"sensor/1/data","qos":1,"retain":false,"size":128}
{"time":"2025-01-28T10:30:02.345678901Z","event":"DISCONNECT","client_id":"client001","username":"admin","reason":"graceful"}
```
