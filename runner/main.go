// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Local Debug Runner
//
// 本地调试运行器，让开发者无需运行完整 AXMQ 即可测试插件
//
// 使用方法：
//   go run runner/main.go -plugin ./my_plugin.so
//   go run runner/main.go -plugin ./my_plugin.so -script testcases.json

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"plugin"
	"strings"

	"github.com/AXMQ-NET/axmq-plugin-sdk/pluginapi"
)

var (
	pluginPath = flag.String("plugin", "", "Path to plugin .so file")
	scriptPath = flag.String("script", "", "Path to test script JSON file (optional)")
	configPath = flag.String("config", "", "Path to plugin config file (optional)")
)

func main() {
	flag.Parse()

	if *pluginPath == "" {
		fmt.Println("Usage: go run runner/main.go -plugin <path/to/plugin.so>")
		os.Exit(1)
	}

	// 加载插件
	plug, err := loadPlugin(*pluginPath)
	if err != nil {
		fmt.Printf("Failed to load plugin: %v\n", err)
		os.Exit(1)
	}
	defer plug.Close()

	// 显示插件信息
	info := plug.Info()
	fmt.Printf("Plugin loaded successfully:\n")
	fmt.Printf("  Name:        %s\n", info.Name)
	fmt.Printf("  Version:     %s\n", info.Version)
	fmt.Printf("  SDK Version: %s\n", info.SDKVersion)
	fmt.Printf("  Go Version:  %s\n", info.GoVersion)
	fmt.Printf("  Build Time:  %s\n", info.BuildTime)
	fmt.Println()

	// 初始化插件
	var config []byte
	if *configPath != "" {
		config, err = os.ReadFile(*configPath)
		if err != nil {
			fmt.Printf("Warning: failed to read config file: %v\n", err)
		}
	}
	if err := plug.Init(config); err != nil {
		fmt.Printf("Plugin initialization failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Plugin initialized.")

	// 如果指定了测试脚本，执行脚本
	if *scriptPath != "" {
		runScript(plug, *scriptPath)
		return
	}

	// 交互式模式
	runInteractive(plug)
}

func loadPlugin(path string) (pluginapi.Plugin, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}

	sym, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, pluginapi.ErrSymbolNotFound
	}

	newFunc, ok := sym.(func() pluginapi.Plugin)
	if !ok {
		return nil, pluginapi.ErrInvalidPluginType
	}

	plug := newFunc()

	// 校验元信息
	info := plug.Info()
	if err := info.Validate(); err != nil {
		return nil, err
	}

	return plug, nil
}

func runInteractive(plug pluginapi.Plugin) {
	fmt.Println("Interactive mode. Type 'help' for available commands.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "help":
			printHelp()
		case "auth":
			handleAuth(plug, parts[1:])
		case "subscribe":
			handleSubscribe(plug, parts[1:])
		case "publish":
			handlePublish(plug, parts[1:])
		case "disconnect":
			handleDisconnect(plug, parts[1:])
		case "exit", "quit":
			fmt.Println("Bye!")
			return
		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  auth <clientID> <username> <password> <ip>")
	fmt.Println("    Test OnAuth hook")
	fmt.Println("  subscribe <clientID> <username> <topic> <qos>")
	fmt.Println("    Test OnSubscribe hook")
	fmt.Println("  publish <clientID> <username> <topic> <payload> <qos> [retain]")
	fmt.Println("    Test OnPublish hook")
	fmt.Println("  disconnect <clientID> <username> [reason]")
	fmt.Println("    Test OnDisconnect hook")
	fmt.Println("  exit / quit")
	fmt.Println("    Exit the runner")
}

func handleAuth(plug pluginapi.Plugin, args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: auth <clientID> <username> <password> <ip>")
		return
	}

	ctx := &pluginapi.AuthContext{
		ClientID: args[0],
		Username: args[1],
		Password: []byte(args[2]),
		IP:       args[3],
	}

	allow, err := plug.OnAuth(ctx)
	if err != nil {
		fmt.Printf("OnAuth error: %v\n", err)
	}
	fmt.Printf("OnAuth result: allow=%v\n", allow)
}

func handleSubscribe(plug pluginapi.Plugin, args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: subscribe <clientID> <username> <topic> <qos>")
		return
	}

	var qos uint8
	fmt.Sscanf(args[3], "%d", &qos)

	ctx := &pluginapi.SubscribeContext{
		ClientID: args[0],
		Username: args[1],
		Topic:    args[2],
		QoS:      qos,
	}

	allow, err := plug.OnSubscribe(ctx)
	if err != nil {
		fmt.Printf("OnSubscribe error: %v\n", err)
	}
	fmt.Printf("OnSubscribe result: allow=%v\n", allow)
}

func handlePublish(plug pluginapi.Plugin, args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: publish <clientID> <username> <topic> <payload> <qos> [retain]")
		return
	}

	var qos uint8
	fmt.Sscanf(args[4], "%d", &qos)

	retain := false
	if len(args) > 5 && (args[5] == "true" || args[5] == "1") {
		retain = true
	}

	ctx := &pluginapi.PublishContext{
		ClientID: args[0],
		Username: args[1],
		Topic:    args[2],
		Payload:  []byte(args[3]),
		QoS:      qos,
		Retain:   retain,
	}

	plug.OnPublish(ctx)
	fmt.Println("OnPublish called (async hook, no return value)")
}

func handleDisconnect(plug pluginapi.Plugin, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: disconnect <clientID> <username> [reason]")
		return
	}

	reason := "graceful"
	if len(args) > 2 {
		reason = args[2]
	}

	ctx := &pluginapi.DisconnectContext{
		ClientID: args[0],
		Username: args[1],
		Reason:   reason,
	}

	plug.OnDisconnect(ctx)
	fmt.Println("OnDisconnect called")
}

// TestCase 测试用例结构
type TestCase struct {
	Name   string          `json:"name"`
	Hook   string          `json:"hook"`
	Input  json.RawMessage `json:"input"`
	Expect struct {
		Allow *bool  `json:"allow,omitempty"`
		Error string `json:"error,omitempty"`
	} `json:"expect"`
}

func runScript(plug pluginapi.Plugin, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Failed to read script file: %v\n", err)
		return
	}

	var cases []TestCase
	if err := json.Unmarshal(data, &cases); err != nil {
		fmt.Printf("Failed to parse script file: %v\n", err)
		return
	}

	passed := 0
	failed := 0

	for i, tc := range cases {
		fmt.Printf("[%d] %s ... ", i+1, tc.Name)

		var result bool
		var resultErr error

		switch tc.Hook {
		case "auth":
			var ctx pluginapi.AuthContext
			json.Unmarshal(tc.Input, &ctx)
			result, resultErr = plug.OnAuth(&ctx)
		case "subscribe":
			var ctx pluginapi.SubscribeContext
			json.Unmarshal(tc.Input, &ctx)
			result, resultErr = plug.OnSubscribe(&ctx)
		case "publish":
			var ctx pluginapi.PublishContext
			json.Unmarshal(tc.Input, &ctx)
			plug.OnPublish(&ctx)
			result = true // OnPublish has no return value
		case "disconnect":
			var ctx pluginapi.DisconnectContext
			json.Unmarshal(tc.Input, &ctx)
			plug.OnDisconnect(&ctx)
			result = true
		default:
			fmt.Printf("SKIP (unknown hook: %s)\n", tc.Hook)
			continue
		}

		// 检查结果
		ok := true
		if tc.Expect.Allow != nil && result != *tc.Expect.Allow {
			ok = false
		}
		if tc.Expect.Error != "" && (resultErr == nil || !strings.Contains(resultErr.Error(), tc.Expect.Error)) {
			ok = false
		}

		if ok {
			fmt.Println("PASS")
			passed++
		} else {
			fmt.Printf("FAIL (got allow=%v, err=%v)\n", result, resultErr)
			failed++
		}
	}

	fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)
}
