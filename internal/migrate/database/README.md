# 数据库初始化说明

## 文件说明

### 00001_init_basic_data.sql
- **用途**: 初始化认证方法和默认支付方式的基础数据
- **租户ID**: 1（默认租户）
- **内容**:
  - 7种认证方法配置
  - 余额支付方式

### 00002_init_tenant_data.sql ⭐ **完整版**
- **用途**: 完整的租户初始化数据（推荐使用）
- **租户ID**: 1（默认租户）
- **内容**:
  - 8种认证方法配置（包括Facebook）
  - 默认余额支付方式
  - 45项系统配置
  - 可选的管理员用户、订阅应用、订阅组配置

## 使用方法

### 方式1: 使用完整初始化脚本（推荐）

```bash
# 进入数据库
mysql -u root -p your_database_name

# 执行完整初始化
source /path/to/00002_init_tenant_data.sql
```

### 方式2: 使用基础初始化脚本

```bash
# 执行基础数据初始化
source /path/to/00001_init_basic_data.sql
```

### 方式3: 使用程序自动初始化

项目启动时会自动检查并创建数据库表结构，但不会自动插入初始数据。建议手动执行SQL文件。

## 初始化数据详情

### 1. 认证方法配置 (proxy_auth_method)

| 方法 | 启用状态 | 说明 |
|------|---------|------|
| email | ✅ 已启用 | SMTP邮件认证 |
| mobile | ❌ 未启用 | 短信认证（阿里云） |
| apple | ❌ 未启用 | Apple Sign In |
| google | ❌ 未启用 | Google OAuth |
| github | ❌ 未启用 | GitHub OAuth |
| facebook | ❌ 未启用 | Facebook OAuth |
| telegram | ❌ 未启用 | Telegram Bot |
| device | ❌ 未启用 | 设备认证 |

### 2. 支付方式 (proxy_payment)

| 名称 | 平台 | 启用状态 |
|------|------|---------|
| Balance | balance | ✅ 已启用 |

### 3. 系统配置 (proxy_system)

#### 站点配置 (site)
- `SiteLogo`: /favicon.svg
- `SiteName`: Perfect Panel
- `SiteDesc`: 站点描述
- `Host`: 站点域名（需配置）
- `Keywords`: Perfect Panel,PPanel
- `CustomHTML`: 自定义HTML

#### 服务条款 (tos)
- `TosContent`: 服务条款内容
- `PrivacyPolicy`: 隐私政策

#### 订阅配置 (subscribe)
- `SingleModel`: false（多订阅模式）
- `SubscribePath`: /api/subscribe
- `SubscribeDomain`: 订阅域名（需配置）
- `PanDomain`: false（不使用泛域名）

#### 验证码配置 (verify)
- `TurnstileSiteKey`: Cloudflare Turnstile密钥
- `TurnstileSecret`: Cloudflare Turnstile密钥
- `EnableLoginVerify`: false
- `EnableRegisterVerify`: false
- `EnableResetPasswordVerify`: false

#### 服务器配置 (server)
- `NodeSecret`: 12345678（节点通信密钥，建议修改）
- `NodePullInterval`: 10秒
- `NodePushInterval`: 60秒
- `NodeMultiplierConfig`: []

#### 邀请配置 (invite)
- `ForcedInvite`: false（不强制邀请码）
- `ReferralPercentage`: 20%
- `OnlyFirstPurchase`: false

#### 注册配置 (register)
- `StopRegister`: false（允许注册）
- `EnableTrial`: false（不启用试用）
- `TrialTime`: 24小时
- `EnableIpRegisterLimit`: false
- `IpRegisterLimit`: 3次
- `IpRegisterLimitDuration`: 64分钟

#### 货币配置 (currency)
- `Currency`: USD
- `CurrencySymbol`: $
- `CurrencyUnit`: USD
- `AccessKey`: 汇率API密钥（需配置）

#### 验证码限制 (verify_code)
- `VerifyCodeExpireTime`: 300秒（5分钟）
- `VerifyCodeLimit`: 15次
- `VerifyCodeInterval`: 60秒

#### 系统信息 (system)
- `Version`: 1.0.0

## 后续配置步骤

### 1. 创建管理员账户

编辑 `00002_init_tenant_data.sql`，取消第4部分的注释：

```sql
INSERT INTO `proxy_user` (`tenant_id`, `password`, ...)
VALUES (1, '$2a$10$...', ...);
```

或通过API创建：
```bash
curl -X POST http://localhost:8080/v1/admin/user \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "is_admin": true
  }'
```

### 2. 配置第三方服务

根据实际需求，更新以下配置：

```sql
-- 配置SMTP邮件服务
UPDATE proxy_auth_method
SET config = '{
    "platform": "smtp",
    "platform_config": {
        "host": "smtp.gmail.com",
        "port": 587,
        "user": "your_email@gmail.com",
        "pass": "your_password",
        "from": "noreply@example.com",
        "ssl": true
    },
    "enable_verify": true,
    "enable_notify": true
}'
WHERE tenant_id = 1 AND method = 'email';

-- 启用邮件认证
UPDATE proxy_auth_method
SET enabled = 1
WHERE tenant_id = 1 AND method = 'email';
```

### 3. 配置订阅应用

取消第5部分注释或执行：

```sql
INSERT INTO `proxy_subscribe_application`
(`tenant_id`, `name`, `icon`, `scheme`, `output_format`, ...)
VALUES
(1, 'Clash', '', 'clash', 'clash', ...),
(1, 'V2rayN', '', 'v2rayn', 'v2ray', ...);
```

### 4. 配置订阅分组

取消第6部分注释或执行：

```sql
INSERT INTO `proxy_subscribe_group`
(`tenant_id`, `name`, `remarks`, `sort`, ...)
VALUES (1, '默认分组', '系统默认订阅分组', 0, ...);
```

## 安全建议

1. **修改默认密钥**
   ```sql
   UPDATE proxy_system
   SET value = 'YOUR_SECURE_SECRET_KEY'
   WHERE tenant_id = 1 AND category = 'server' AND key = 'NodeSecret';
   ```

2. **启用验证码**
   ```sql
   UPDATE proxy_system
   SET value = 'true'
   WHERE tenant_id = 1 AND category = 'verify'
     AND key IN ('EnableLoginVerify', 'EnableRegisterVerify');
   ```

3. **配置IP注册限制**
   ```sql
   UPDATE proxy_system
   SET value = 'true'
   WHERE tenant_id = 1 AND category = 'register'
     AND key = 'EnableIpRegisterLimit';
   ```

## 多租户支持

如需创建新租户，复制并修改SQL文件中的tenant_id：

```sql
-- 租户2的初始化数据
INSERT INTO `proxy_auth_method` (`tenant_id`, `method`, ...)
VALUES (2, 'email', ...), ...;

INSERT INTO `proxy_payment` (`tenant_id`, `name`, ...)
VALUES (2, 'Balance', ...);

INSERT INTO `proxy_system` (`tenant_id`, `category`, `key`, ...)
VALUES (2, 'site', 'SiteName', 'Tenant 2 Site', ...);
```

## 验证初始化

执行以下SQL验证初始化是否成功：

```sql
-- 检查认证方法
SELECT method, enabled FROM proxy_auth_method WHERE tenant_id = 1;

-- 检查支付方式
SELECT name, enable FROM proxy_payment WHERE tenant_id = 1;

-- 检查系统配置
SELECT category, COUNT(*) as count
FROM proxy_system
WHERE tenant_id = 1
GROUP BY category;

-- 预期结果：
-- site: 6项
-- tos: 2项
-- ad: 1项
-- subscribe: 4项
-- verify: 5项
-- server: 4项
-- invite: 3项
-- register: 8项
-- currency: 4项
-- verify_code: 3项
-- system: 1项
-- 总计: 41项
```

## 故障排查

### 问题1: 重复键错误
```
ERROR 1062 (23000): Duplicate entry '1-site-SiteName' for key 'proxy_system.proxy_system_tenant_id_category_key'
```

**解决方案**: SQL使用了 `ON DUPLICATE KEY UPDATE`，重复执行是安全的。如需重新初始化：

```sql
DELETE FROM proxy_system WHERE tenant_id = 1;
DELETE FROM proxy_auth_method WHERE tenant_id = 1;
DELETE FROM proxy_payment WHERE tenant_id = 1;
-- 然后重新执行初始化SQL
```

### 问题2: 外键约束错误

**解决方案**: 确保先创建表结构，再执行初始化数据：

```bash
# 1. 启动应用让ent自动创建表
go run ./cmd/kratos-service

# 2. 执行初始化SQL
mysql -u root -p your_database < 00002_init_tenant_data.sql
```

## 相关文档

- [Ent Schema 文档](../../ent/schema/)
- [配置管理文档](../../../docs/configuration.md)
- [多租户架构文档](../../../docs/multi-tenant.md)
