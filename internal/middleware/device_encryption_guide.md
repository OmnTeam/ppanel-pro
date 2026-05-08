# 设备中间件 AES 加解密使用指南

## 概述

设备中间件现在已完整实现AES加解密功能，支持对设备请求和响应的自动加密解密处理。

## 🔐 加密算法

- **算法**: AES-256-CBC
- **填充**: PKCS7
- **密钥派生**: MD5哈希
- **IV生成**: MD5(nonce + secret)

## 📡 支持的加密位置

### 1. URL查询参数加密
```http
GET /api/device?data=encrypted_base64&time=nonce
```

### 2. 请求体加密
```json
{
  "data": "encrypted_base64",
  "time": "nonce"
}
```

### 3. 响应体加密
```json
{
  "data": "encrypted_base64",
  "time": "nonce"
}
```

## ⚙️ 环境变量配置

```bash
# 启用设备中间件
export DEVICE_MIDDLEWARE_ENABLE="true"

# 设置设备加密密钥（与客户端保持一致）
export DEVICE_SECURITY_SECRET="your-device-secret-key-here"
```

## 🚀 使用方法

### 1. 在服务中启用设备中间件

```go
import "github.com/OmnTeam/npanel-pro/internal/middleware"

// 创建服务上下文
svc := &middleware.ServiceContext{
    Config:    bootstrapConfig,
    Redis:     redisClient,
    UserModel: userService,
}

// 在HTTP服务器中添加设备中间件
http.Middleware(
    svc.CORS(),
    svc.Logger(),
    svc.Trace(),
    svc.Auth(),
    svc.Device(), // 设备中间件
    svc.Server(),
)
```

### 2. 客户端请求加密

#### 加密URL参数
```bash
# 原始参数
param1=value1&param2=value2

# 加密后
data=ENCRYPTED_BASE64&time=NONCE
```

#### 加密请求体
```json
// 原始请求体
{
  "username": "test",
  "password": "123456"
}

// 加密后请求体
{
  "data": "ENCRYPTED_BASE64",
  "time": "NONCE"
}
```

### 3. 请求头设置

```http
# 设备类型请求
Login-Type: device

# 要求响应加密
X-Device-Encrypt: true

# 其他设备信息
User-Agent: DeviceClient/1.0
```

### 4. 服务端处理解密数据

```go
func (s *DeviceService) HandleDeviceRequest(ctx context.Context, req *pb.DeviceRequest) (*pb.DeviceResponse, error) {
    // 获取解密后的查询参数
    if queryData, ok := middleware.GetDecryptedQueryData(ctx); ok {
        logger.WithContext(ctx).Info("Decrypted query data", logger.Field("data", queryData))
    }

    // 获取解密后的请求体数据
    if bodyData, ok := middleware.GetDecryptedBodyData(ctx); ok {
        logger.WithContext(ctx).Info("Decrypted body data", logger.Field("data", bodyData))

        // 将解密数据转换为具体类型
        if deviceReq, ok := bodyData.(map[string]interface{}); ok {
            username := deviceReq["username"]
            password := deviceReq["password"]
            // 处理业务逻辑...
        }
    }

    // 返回响应（会自动加密）
    return &pb.DeviceResponse{
        Code:    200,
        Message: "success",
        Data:    "response data",
    }, nil
}
```

## 🔄 完整的加解密流程

### 请求流程
1. **客户端**: 使用设备密钥加密数据
2. **发送请求**: 包含加密的data和时间戳time
3. **中间件解密**: 自动解密URL参数和请求体
4. **业务处理**: 获取解密后的数据进行业务处理
5. **返回响应**: 响应数据自动加密

### 响应流程
1. **服务端**: 处理业务逻辑生成响应
2. **中间件加密**: 检测到`X-Device-Encrypt: true`自动加密响应
3. **返回加密响应**: 包含加密的data和时间戳time
4. **客户端解密**: 使用相同的密钥和时间戳解密响应

## 🔧 辅助函数

### 获取解密数据
```go
// 获取解密后的查询参数
queryData, hasQuery := middleware.GetDecryptedQueryData(ctx)

// 获取解密后的请求体数据
bodyData, hasBody := middleware.GetDecryptedBodyData(ctx)

// 获取原始请求数据
originalReq, hasOriginal := middleware.GetOriginalRequest(ctx)

// 统一解析设备请求
parsedReq, err := middleware.ParseDeviceRequest(ctx, req)
```

## 📝 示例代码

### 客户端加密示例（伪代码）
```python
import base64
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad
import hashlib
import time

def encrypt_data(data, secret):
    # 生成密钥
    key = hashlib.md5(secret.encode()).digest()

    # 生成时间戳作为nonce
    nonce = str(int(time.time() * 1000000))

    # 生成IV
    iv = hashlib.md5((nonce + secret).encode()).digest()[:16]

    # 加密数据
    cipher = AES.new(key, AES.MODE_CBC, iv)
    padded_data = pad(data.encode(), AES.block_size)
    encrypted = cipher.encrypt(padded_data)

    # 返回Base64编码的加密数据和时间戳
    return base64.b64encode(encrypted).decode(), nonce

# 使用示例
secret = "your-device-secret-key"
original_data = '{"username":"test","password":"123456"}'
encrypted_data, nonce = encrypt_data(original_data, secret)

# 构造请求
request = {
    "data": encrypted_data,
    "time": nonce
}
```

### Go服务端处理示例
```go
type DeviceLoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func (s *DeviceService) Login(ctx context.Context, req *pb.DeviceLoginRequest) (*pb.DeviceLoginResponse, error) {
    // 获取解密后的数据
    decryptedData, ok := middleware.GetDecryptedBodyData(ctx)
    if !ok {
        return nil, errors.BadRequest("BAD_REQUEST", "No encrypted data found")
    }

    // 解析解密后的数据
    dataBytes, err := json.Marshal(decryptedData)
    if err != nil {
        return nil, errors.InternalServer("INTERNAL_ERROR", "Failed to marshal decrypted data")
    }

    var loginReq DeviceLoginRequest
    if err := json.Unmarshal(dataBytes, &loginReq); err != nil {
        return nil, errors.BadRequest("BAD_REQUEST", "Invalid request format")
    }

    // 验证用户名密码
    if err := s.validateUser(loginReq.Username, loginReq.Password); err != nil {
        return nil, errors.Unauthorized("UNAUTHORIZED", "Invalid credentials")
    }

    // 生成token
    token, err := s.generateToken(loginReq.Username)
    if err != nil {
        return nil, errors.InternalServer("INTERNAL_ERROR", "Failed to generate token")
    }

    // 返回响应（会自动加密）
    return &pb.DeviceLoginResponse{
        Code:    200,
        Message: "Login successful",
        Data: map[string]interface{}{
            "token": token,
            "expires_in": 3600,
        },
    }, nil
}
```

## ⚠️ 注意事项

1. **密钥安全**: 确保`DEVICE_SECURITY_SECRET`与客户端保持一致
2. **时间同步**: 客户端和服务端的时间差不能太大
3. **错误处理**: 加密失败时会返回详细的错误信息
4. **性能考虑**: 加解密会带来一定的性能开销
5. **调试模式**: 可以通过日志查看加解密过程的详细信息

## 🔍 调试和监控

### 查看解密日志
```bash
# 设置日志级别为DEBUG
export LOG_LEVEL=debug

# 启动服务
go run ./cmd/npanel-pro -conf=configs/config.yaml
```

### 监控加解密性能
```go
// 在中间件中已内置性能监控
// 可以通过日志查看加解密的耗时信息
logger.WithContext(ctx).Debug("[DeviceMiddleware] Decrypt completed",
    logger.Field("duration", duration))
```

## ✅ 功能状态

- ✅ URL参数解密
- ✅ 请求体解密
- ✅ 响应体加密
- ✅ AES-256-CBC加密算法
- ✅ PKCS7填充
- ✅ 动态IV生成
- ✅ 完整的错误处理
- ✅ 详细的日志记录
- ✅ 性能优化

**设备中间件AES加解密功能已完全实现并可投入使用！** 🎉