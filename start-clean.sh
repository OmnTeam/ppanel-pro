#!/bin/bash

# 启动应用程序并过滤掉 Redis 兼容性警告
echo "Starting PPanel Pro with Redis warnings filtered..."

# 运行应用程序，过滤掉特定的 Redis 警告信息
go run ./cmd/ppanel-pro -conf=configs/config.yaml 2>&1 | grep -v "maint_notifications disabled" | grep -v "ERR unknown subcommand 'maint_notifications'"

echo "Application stopped."