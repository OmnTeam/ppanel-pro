#!/bin/bash

# 启动应用程序并过滤掉 Redis 维护通知警告
go run ./cmd/ppanel-pro -conf=configs/config.yaml 2>&1 | grep -v "maint_notifications disabled"