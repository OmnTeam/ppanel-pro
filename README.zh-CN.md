# NPanel Pro

<p align="center">
  <strong>基于 Go 与 Kratos 构建的专业化面板管理系统后端。</strong>
</p>

<p align="center">
  <a href="./README.md">English</a> | 简体中文
</p>

## 项目介绍

NPanel Pro 是一个面向订阅平台及相关业务系统的专业化面板管理后端。项目同时提供 HTTP 与 gRPC API，使用 Proto 文件定义服务契约，并集成 MySQL、Redis、Ent ORM 与异步任务处理能力，适合生产环境下的可扩展部署。

它覆盖用户管理、认证、订阅、订单、支付、公告、工单、文档、营销与系统管理等核心业务场景。

## 项目架构

### 架构概览

- `cmd/ppanel-pro`：应用启动入口，负责配置加载、日志初始化、Wire 依赖注入与服务启动。
- `api`：Proto 契约与生成代码，按 `admin`、`auth`、`public`、`server` 等领域拆分。
- `internal/service`：接口服务层，负责对外暴露 API 并编排请求处理流程。
- `internal/biz`：核心业务层，承载用户、订阅、订单、支付、工单、营销与系统功能等领域逻辑。
- `internal/data`：数据访问层，负责 MySQL、Redis 与持久化实现。
- `internal/server`：HTTP/gRPC 服务注册、中间件、兼容路由与跨域配置。
- `internal/queue`：异步任务与定时处理，如邮件、短信、订单关闭、流量统计、订阅检查等。
- `ent`：Ent Schema 与 ORM 相关生成代码。
- `pkg`：通用基础设施能力封装，如支付、邮件、短信、JWT、缓存、模板、追踪与工具库。
- `configs`：运行配置文件。

### 分层示意

```text
Client / Admin / Public Apps
            |
      HTTP / gRPC APIs
            |
     internal/service
            |
       internal/biz
            |
 internal/data + ent ORM
            |
      MySQL / Redis

Async Processing: internal/queue
Shared Utilities: pkg/*
Contracts: api/* proto
Bootstrap: cmd/ppanel-pro
```

## 功能特性

- 用户与认证：支持邮箱、手机号、验证码、找回密码、管理员登录等流程。
- 订阅与用户管理：支持套餐/分组、设备、流量、用户资料与状态管理。
- 订单与支付：支持优惠券、兑换码，以及 Stripe、Alipay F2F、EPay、CryptoSaaS、余额支付等方式。
- 内容与运营：支持公告、文档、广告、营销与应用配置管理。
- 工单与后台管理：支持工单、日志、系统设置、服务端管理与后台控制台能力。
- 异步任务：内置邮件、短信、流量统计、订阅检查与订单生命周期处理任务。
- 双协议接口：同时提供 HTTP 与 gRPC，便于多端接入与内部服务调用。
- 可扩展基础设施：基于 Proto、Wire、Ent、Redis、结构化日志与中间件机制构建。

## 快速开始

### 安装 Kratos

```bash
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```

### 生成 API 与辅助文件

```bash
# 下载并更新依赖
make init

# 生成 API 文件（pb.go、http、grpc、validate、swagger）
make api

# 生成全部文件
make all
```

### Wire 初始化

```bash
# 安装 wire
go install github.com/google/wire/cmd/wire@latest

# 生成 wire 代码
cd cmd/ppanel-pro
wire
```

### 构建与运行

```bash
go generate ./...
go build -o ./bin/ppanel-pro ./cmd/ppanel-pro
./bin/ppanel-pro -conf ./configs
```

### Docker

```bash
# 构建
docker build -t npanel-pro .

# 运行（默认配置暴露 HTTP 8081 与 gRPC 9012）
docker run --rm -p 8081:8081 -p 9012:9012 -v </path/to/your/configs>:/data/conf npanel-pro
```
