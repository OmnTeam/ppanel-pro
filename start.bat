@echo off
REM 启动应用程序并过滤掉 Redis 维护通知警告
go run ./cmd/ppanel-pro -conf=configs/config.yaml 2>&1 | findstr /v "maint_notifications disabled"