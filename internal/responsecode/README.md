# Proxy Service 响应码使用指南

## 概述

本模块实现了基于 SaaS Response Code Design v4.0 规范的响应码系统，用于统一错误处理和响应格式。

**模块码：060** (代理服务)
**文件类型码：003** (业务逻辑)

## 响应码格式

```
分类码(1位) + 模块码(3位) + 文件类型码(3位) + 序号(3位)
例如：2060003000
  2    - 成功分类
  060  - 代理服务模块
  003  - 业务逻辑文件
  000  - 序号
```

## 响应码分类

| 分类码 | 范围 | 说明 | HTTP状态码 |
|--------|------|------|------------|
| 2 | 2060003000-2060003999 | 成功响应 | 200 |
| 3 | 3060003000-3060003999 | 业务错误 | 400 |
| 4 | 4060003000-4060003999 | 权限错误 | 401/403 |
| 5 | 5060003000-5060003999 | 系统错误 | 500/503 |

## 使用方法

### 1. 在 Service 层返回成功响应

```go
package service

import (
    "context"
    v1 "github.com/OmnTeam/npanel-pro/api/coupon/v1"
    "github.com/OmnTeam/npanel-pro/internal/biz"
    "github.com/OmnTeam/npanel-pro/internal/responsecode"
)

func (s *CouponService) CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CouponReply, error) {
    coupon, err := s.uc.CreateCoupon(ctx, req.TenantId, req.Name, req.Code, ...)
    if err != nil {
        return nil, err  // 错误由 Biz 层处理
    }

    return &v1.CouponReply{
        Coupon: s.convertToProto(coupon),
    }, nil
}
```

### 2. 在 Biz 层处理业务错误

```go
package biz

import (
    "github.com/OmnTeam/npanel-pro/internal/responsecode"
)

// 示例1: 参数验证错误
func (uc *CouponUsecase) GetCoupon(ctx context.Context, tenantID, couponID int64) (*Coupon, error) {
    if tenantID <= 0 {
        return nil, responsecode.NewKratosError(
            responsecode.ErrInvalidTenantID,
            "租户ID必须大于0",
        )
    }

    if couponID <= 0 {
        return nil, responsecode.ErrInvalidParam("coupon_id")
    }

    // ... 业务逻辑
}

// 示例2: 资源不存在
func (uc *CouponUsecase) GetCouponByCode(ctx context.Context, tenantID int64, code string) (*Coupon, error) {
    coupon, err := uc.repo.GetCouponByCode(ctx, tenantID, code)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, responsecode.NewKratosError(
                responsecode.ErrCouponNotFound,
                "优惠券不存在",
            )
        }
        return nil, responsecode.ErrDatabaseOperation("查询")
    }
    return coupon, nil
}

// 示例3: 业务逻辑错误
func (uc *CouponUsecase) ValidateCoupon(ctx context.Context, tenantID int64, code string) (*Coupon, error) {
    coupon, err := uc.repo.GetCouponByCode(ctx, tenantID, code)
    if err != nil {
        return nil, err
    }

    // 检查是否启用
    if !coupon.Enable {
        return nil, responsecode.NewKratosError(
            responsecode.ErrCouponNotAvailable,
            "优惠券不可用",
        )
    }

    // 检查是否过期
    now := time.Now()
    if now.Before(coupon.StartTime) || now.After(coupon.ExpireTime) {
        return nil, responsecode.NewKratosError(
            responsecode.ErrCouponExpired,
            "优惠券已过期",
        )
    }

    // 检查库存
    if coupon.Count > 0 && coupon.UsedCount >= coupon.Count {
        return nil, responsecode.NewKratosError(
            responsecode.ErrCouponUsedUp,
            "优惠券已用完",
        )
    }

    return coupon, nil
}

// 示例4: 资源已存在
func (uc *UserUsecase) CreateUser(ctx context.Context, email string) (*User, error) {
    // 检查邮箱是否已存在
    exists, err := uc.repo.CheckEmailExists(ctx, email)
    if err != nil {
        return nil, responsecode.ErrDatabaseOperation("查询")
    }
    if exists {
        return nil, responsecode.NewKratosError(
            responsecode.ErrDuplicateEmail,
            "该邮箱已被注册",
        )
    }

    // ... 创建用户
}
```

### 3. 在 Data 层处理数据库错误

```go
package data

import (
    "github.com/OmnTeam/npanel-pro/internal/responsecode"
)

// 示例1: 数据库查询错误
func (r *couponRepo) GetCoupon(ctx context.Context, tenantID, id int64) (*biz.Coupon, error) {
    po, err := r.data.db.Coupon.
        Query().
        Where(
            coupon.TenantID(tenantID),
            coupon.ID(id),
        ).
        Only(ctx)

    if err != nil {
        if ent.IsNotFound(err) {
            // 不存在应该返回业务错误，由上层 Biz 处理
            return nil, err
        }
        // 数据库错误
        return nil, responsecode.NewKratosError(
            responsecode.ErrDatabaseQuery,
            "查询优惠券失败",
        )
    }

    return r.convertToDomain(po), nil
}

// 示例2: 数据库插入错误
func (r *couponRepo) CreateCoupon(ctx context.Context, coupon *biz.Coupon) (*biz.Coupon, error) {
    po, err := r.data.db.Coupon.
        Create().
        SetTenantID(coupon.TenantID).
        SetName(coupon.Name).
        SetCode(coupon.Code).
        // ... 其他字段
        Save(ctx)

    if err != nil {
        // 检查是否是唯一约束冲突
        if ent.IsConstraintError(err) {
            return nil, responsecode.NewKratosError(
                responsecode.ErrCouponAlreadyExists,
                "优惠券代码已存在",
            )
        }
        return nil, responsecode.NewKratosError(
            responsecode.ErrDatabaseInsert,
            "创建优惠券失败",
        )
    }

    return r.convertToDomain(po), nil
}

// 示例3: 数据库事务错误
func (r *orderRepo) CreateOrder(ctx context.Context, order *biz.Order) (*biz.Order, error) {
    tx, err := r.data.db.Tx(ctx)
    if err != nil {
        return nil, responsecode.NewKratosError(
            responsecode.ErrDatabaseTransaction,
            "开启事务失败",
        )
    }

    defer func() {
        if v := recover(); v != nil {
            tx.Rollback()
        }
    }()

    // 执行事务操作...

    if err := tx.Commit(); err != nil {
        return nil, responsecode.NewKratosError(
            responsecode.ErrDatabaseTransaction,
            "提交事务失败",
        )
    }

    return order, nil
}
```

### 4. 便捷方法

```go
// 快捷创建常见错误
func (uc *UserUsecase) GetUser(ctx context.Context, id int64) (*User, error) {
    user, err := uc.repo.GetUser(ctx, id)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, responsecode.ErrNotFound("用户")
        }
        return nil, err
    }
    return user, nil
}

// 参数验证错误
if name == "" {
    return nil, responsecode.ErrMissingParam("name")
}

if !isValidEmail(email) {
    return nil, responsecode.ErrInvalidParam("email")
}

// 数据库操作错误
if err != nil {
    return nil, responsecode.ErrDatabaseOperation("更新")
}

// 缓存操作错误
if err != nil {
    return nil, responsecode.ErrCacheOperation("设置")
}

// 认证和授权错误
return nil, responsecode.ErrUnauthorized()
return nil, responsecode.ErrForbidden()
```

## 错误码列表

### 成功码 (2060003XXX)

| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 2060003000 | UserCreated | 用户创建成功 |
| 2060003001 | UserUpdated | 用户更新成功 |
| 2060003100 | SubscribeCreated | 订阅创建成功 |
| 2060003200 | OrderCreated | 订单创建成功 |
| 2060003300 | PaymentCreated | 支付方式创建成功 |
| 2060003400 | ServerCreated | 服务器创建成功 |
| 2060003500 | NodeCreated | 节点创建成功 |
| 2060003600 | CouponCreated | 优惠券创建成功 |

### 业务错误 (3060003XXX)

#### 参数验证错误 (3060003000-3060003099)
| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 3060003000 | ErrInvalidUserID | 无效的用户ID |
| 3060003001 | ErrInvalidTenantID | 无效的租户ID |
| 3060003007 | ErrInvalidCouponCode | 无效的优惠券码 |
| 3060003008 | ErrMissingRequiredParam | 缺少必需参数 |

#### 数据不存在错误 (3060003100-3060003199)
| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 3060003100 | ErrUserNotFound | 用户不存在 |
| 3060003101 | ErrOrderNotFound | 订单不存在 |
| 3060003106 | ErrCouponNotFound | 优惠券不存在 |

#### 数据冲突错误 (3060003200-3060003299)
| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 3060003200 | ErrUserAlreadyExists | 用户已存在 |
| 3060003206 | ErrCouponAlreadyExists | 优惠券已存在 |
| 3060003207 | ErrDuplicateEmail | 邮箱已存在 |

#### 业务逻辑错误 (3060003300-3060003399)
| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 3060003300 | ErrOrderCannotCancel | 订单不能取消 |
| 3060003302 | ErrCouponExpired | 优惠券已过期 |
| 3060003304 | ErrCouponUsedUp | 优惠券已用完 |
| 3060003306 | ErrInsufficientBalance | 余额不足 |

### 权限错误 (4060003XXX)

#### 认证错误 (4060003000-4060003099)
| 错误码 | 常量名 | 说明 | HTTP |
|--------|--------|------|------|
| 4060003000 | ErrMissingAuthToken | 缺少认证令牌 | 401 |
| 4060003001 | ErrInvalidAuthToken | 无效的认证令牌 | 401 |
| 4060003002 | ErrAuthTokenExpired | 认证令牌已过期 | 401 |
| 4060003005 | ErrPasswordIncorrect | 密码错误 | 401 |

#### 授权错误 (4060003100-4060003199)
| 错误码 | 常量名 | 说明 | HTTP |
|--------|--------|------|------|
| 4060003100 | ErrPermissionDenied | 权限被拒绝 | 403 |
| 4060003101 | ErrInsufficientPermission | 权限不足 | 403 |
| 4060003104 | ErrTenantAccessDenied | 租户访问被拒绝 | 403 |

### 系统错误 (5060003XXX)

#### 数据库错误 (5060003000-5060003099)
| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 5060003000 | ErrDatabaseConnection | 数据库连接失败 |
| 5060003001 | ErrDatabaseQuery | 数据库查询失败 |
| 5060003002 | ErrDatabaseUpdate | 数据库更新失败 |
| 5060003005 | ErrDatabaseTransaction | 数据库事务失败 |

#### 缓存错误 (5060003100-5060003199)
| 错误码 | 常量名 | 说明 |
|--------|--------|------|
| 5060003100 | ErrCacheConnection | 缓存连接失败 |
| 5060003101 | ErrCacheGet | 缓存获取失败 |
| 5060003102 | ErrCacheSet | 缓存设置失败 |

## 最佳实践

1. **分层错误处理**
   - Data 层：处理数据库和缓存错误
   - Biz 层：处理业务逻辑错误和参数验证
   - Service 层：直接传递错误，不做处理

2. **错误信息规范**
   - 使用中文错误消息，用户友好
   - 错误消息应明确指出问题所在
   - 避免暴露系统内部实现细节

3. **错误码选择**
   - 优先使用已定义的错误码
   - 新增错误码应遵循规范分类
   - 保持错误码的语义一致性

4. **错误转换**
   - Data 层的 NotFound 错误应传递给 Biz 层处理
   - 数据库约束错误应转换为业务错误
   - 系统错误应记录日志但不暴露细节

## 响应格式

### 成功响应
```json
{
  "code": 2060003600,
  "message": "优惠券创建成功",
  "data": {
    "id": 1,
    "name": "新用户优惠券",
    "code": "NEW2025"
  }
}
```

### 错误响应
```json
{
  "code": 3060003106,
  "message": "优惠券不存在",
  "reason": "COUPON_NOT_FOUND"
}
```

## 参考文档

- [SaaS Response Code Design v4.0](/Doc/Development/response-code-design-v4.md)
- [Permission Service 错误处理](/services/permission-service/internal/responsecode/)