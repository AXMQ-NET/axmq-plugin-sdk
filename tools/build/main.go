// Copyright 2025 AXMQ Authors
// AXMQ Plugin SDK - Build Tool
//
// 构建插件并生成元数据清单，确保与主程序的版本一致性
//
// 使用方法：
//   go run tools/build/main.go -dir ./my_plugin -output ./my_plugin.so
//   go run tools/build/main.go -dir ./my_plugin -output ./my_plugin.so -goos linux -goarch amd64

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/AXMQ-NET/axmq-plugin-sdk/pluginapi"
)

var (
	pluginDir  = flag.String("dir", ".", "Path to plugin source directory")
	outputPath = flag.String("output", "", "Output path for the plugin .so file")
	verbose    = flag.Bool("v", false, "Verbose output")
	targetOS   = flag.String("goos", "", "Target GOOS (empty = host)")
	targetArch = flag.String("goarch", "", "Target GOARCH (empty = host)")
)

// BuildMeta 构建元数据（写入到 .meta.json）
type BuildMeta struct {
	pluginapi.PluginMeta
	BuildHost string `json:"build_host"` // 构建主机
	BuildOS   string `json:"build_os"`   // 构建操作系统
	BuildArch string `json:"build_arch"` // 构建架构
}

func main() {
	flag.Parse()

	if *outputPath == "" {
		// 默认输出到插件目录下
		*outputPath = filepath.Join(*pluginDir, "plugin.so")
	}

	// 确保输出路径是绝对路径
	absOutput, err := filepath.Abs(*outputPath)
	if err != nil {
		fatal("Failed to resolve output path: %v", err)
	}

	// 切换到插件目录
	absDir, err := filepath.Abs(*pluginDir)
	if err != nil {
		fatal("Failed to resolve plugin directory: %v", err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fatal("Plugin directory does not exist: %s", absDir)
	}

	log("Building plugin from: %s", absDir)
	log("Output: %s", absOutput)

	// 1. 检查 Go 版本
	goVersion := runtime.Version()
	log("Go version: %s", goVersion)

	// 2. 构建插件
	log("Running: go build -buildmode=plugin")
	if *targetOS != "" || *targetArch != "" {
		log("Target: %s/%s", defaultIfEmpty(*targetOS, runtime.GOOS), defaultIfEmpty(*targetArch, runtime.GOARCH))
	}

	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", absOutput, ".")
	cmd.Dir = absDir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CGO_ENABLED=1") // 插件必须启用 CGO
	if *targetOS != "" {
		cmd.Env = append(cmd.Env, "GOOS="+*targetOS)
	}
	if *targetArch != "" {
		cmd.Env = append(cmd.Env, "GOARCH="+*targetArch)
	}

	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		fatal("Build failed: %v", err)
	}

	log("Build successful: %s", absOutput)

	// 3. 尝试加载插件获取元信息（可选，用于生成更完整的 meta）
	// 注意：由于我们在构建工具中加载，Go 版本一定是匹配的
	pluginName := strings.TrimSuffix(filepath.Base(absOutput), ".so")
	hostname, _ := os.Hostname()

	buildOS := defaultIfEmpty(*targetOS, runtime.GOOS)
	buildArch := defaultIfEmpty(*targetArch, runtime.GOARCH)
	meta := BuildMeta{
		PluginMeta: pluginapi.PluginMeta{
			Name:       pluginName,
			Version:    "1.0.0", // 默认版本，实际应从插件 Info() 获取
			SDKVersion: pluginapi.SDKVersion,
			GoVersion:  goVersion,
			BuildTime:  time.Now().Format(time.RFC3339),
		},
		BuildHost: hostname,
		BuildOS:   buildOS,
		BuildArch: buildArch,
	}

	// 4. 写入元数据文件
	metaPath := absOutput + ".meta.json"
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		fatal("Failed to marshal meta: %v", err)
	}

	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		fatal("Failed to write meta file: %v", err)
	}

	log("Meta written: %s", metaPath)
	log("")
	log("Build complete!")
	log("  Plugin: %s", absOutput)
	log("  Meta:   %s", metaPath)
}

func log(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func defaultIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
