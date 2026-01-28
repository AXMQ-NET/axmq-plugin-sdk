# Auth Plugin Example

自定义认证插件示例，演示如何实现 OnAuth 钩子对接外部用户系统。

## 构建

```bash
cd examples/auth_plugin
go run ../../tools/build/main.go -dir . -output ./auth_plugin.so
```

## 本地测试

```bash
go run ../../runner/main.go -plugin ./auth_plugin.so
```

交互式命令：
```
> auth client001 admin secret 192.168.1.1
OnAuth result: allow=true

> auth client002 admin wrong 192.168.1.2
OnAuth result: allow=false

> subscribe client001 admin $SYS/broker/stats 0
OnSubscribe result: allow=true

> subscribe client002 guest $SYS/broker/stats 0
OnSubscribe result: allow=false
```

## 配置

创建 `config.json`:

```json
{
  "users": {
    "user1": "pass1",
    "user2": "pass2"
  }
}
```

启动时指定配置：
```bash
go run ../../runner/main.go -plugin ./auth_plugin.so -config ./config.json
```

## 部署到 AXMQ

```bash
cp auth_plugin.so /path/to/axmq/plugins/
cp auth_plugin.so.meta.json /path/to/axmq/plugins/
```
