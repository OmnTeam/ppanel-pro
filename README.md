# NPanel Pro

<p align="center">
  <strong>A professional panel management backend built with Go and Kratos.</strong>
</p>

<p align="center">
  English | <a href="./README.zh-CN.md">简体中文</a>
</p>

## Introduction

NPanel Pro is a professional panel management backend designed for subscription platforms and related business systems. It provides HTTP and gRPC APIs, uses Proto files as service contracts, and integrates MySQL, Redis, Ent ORM, and asynchronous job processing for scalable production deployments.

It targets scenarios such as user management, authentication, subscriptions, orders, payments, announcements, tickets, documents, marketing, and system administration.

## Architecture

### Overview

- `cmd/ppanel-pro`: application entrypoint for config loading, logging bootstrap, Wire dependency injection, and server startup.
- `api`: Proto contracts and generated code, organized by domains such as `admin`, `auth`, `public`, and `server`.
- `internal/service`: API service layer that exposes endpoints and orchestrates request handling.
- `internal/biz`: core domain logic for users, subscriptions, orders, payments, tickets, marketing, and system features.
- `internal/data`: data access layer for MySQL, Redis, and persistence implementations.
- `internal/server`: HTTP/gRPC server registration, middleware, compatibility routes, and CORS handling.
- `internal/queue`: asynchronous jobs and scheduled handlers for email, SMS, order closing, traffic statistics, subscription checks, and more.
- `ent`: Ent schemas and generated ORM artifacts.
- `pkg`: shared infrastructure packages such as payment, email, SMS, JWT, cache, templating, tracing, and utilities.
- `configs`: runtime configuration files.

### Layered View

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

## Features

- User and authentication flows with email, phone, verification code, password reset, and admin login.
- Subscription and user management for plans/groups, devices, traffic, user profiles, and status control.
- Order and payment support for coupons, redemption codes, Stripe, Alipay F2F, EPay, CryptoSaaS, and balance payment.
- Content and operations modules for announcements, documents, ads, marketing, and application configuration.
- Ticketing and administration capabilities for logs, system settings, server management, and admin console operations.
- Built-in asynchronous processing for email, SMS, traffic statistics, subscription checks, and order lifecycle jobs.
- Dual-protocol API exposure through both HTTP and gRPC.
- Extensible infrastructure based on Proto, Wire, Ent, Redis, structured logging, and middleware.

## Quick Start

### Install Kratos

```bash
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
```

### Generate API and auxiliary files

```bash
# Download and update dependencies
make init

# Generate API files (pb.go, http, grpc, validate, swagger)
make api

# Generate all files
make all
```

### Wire initialization

```bash
# Install wire
go install github.com/google/wire/cmd/wire@latest

# Generate wire code
cd cmd/ppanel-pro
wire
```

### Build and run

```bash
go generate ./...
go build -o ./bin/ppanel-pro ./cmd/ppanel-pro
./bin/ppanel-pro -conf ./configs
```

One-off run without hot reload:

```bash
go run ./cmd/ppanel-pro -conf ./configs
```

Run the package `./cmd/ppanel-pro`, not `main.go` alone (Wire’s `wireApp` lives in `wire_gen.go`).

### Local development (Air)

```bash
go install github.com/air-verse/air@latest
air
```

Requires `configs/config.yaml` and reachable MySQL/Redis. Edits under `internal/` and `pkg/` trigger rebuild and restart. Regenerate Proto/Wire/Ent (`make api`, `wire`, `go generate`) before relying on hot reload for those changes.

### Docker

```bash
# Build
docker build -t npanel-pro .

# Run (default config exposes HTTP 8081 and gRPC 9012)
docker run --rm -p 8081:8081 -p 9012:9012 -v </path/to/your/configs>:/data/conf npanel-pro
```
