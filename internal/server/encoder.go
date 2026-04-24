package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ErrorResponse 统一错误响应结构
type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// CustomErrorEncoder 自定义错误编码器
// 所有错误统一返回 HTTP 200，业务错误码在响应体的 code 字段中
func CustomErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	// 默认错误码和消息
	errorCode := 500
	errorMessage := "Internal Server Error"

	// 尝试从 Kratos 错误中提取信息
	if se := errors.FromError(err); se != nil {
		// 从 metadata 中提取 custom_code
		if customCode, ok := se.Metadata["custom_code"]; ok {
			if code, parseErr := strconv.Atoi(customCode); parseErr == nil {
				errorCode = code
			}
		}
		// 使用错误消息
		if se.Message != "" {
			errorMessage = se.Message
		}
	}

	// 构建响应
	response := ErrorResponse{
		Code:    errorCode,
		Message: errorMessage,
	}

	// 序列化响应
	data, err := json.Marshal(response)
	if err != nil {
		// 如果序列化失败，返回纯文本错误
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Internal Server Error"))
		return
	}

	// 返回 JSON 响应，HTTP 状态码统一为 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// 必需字段列表（即使为空也要返回）
var requiredFields = map[string]bool{
	// 基础响应字段
	"code":    true,
	"message": true,
	"data":    true,
	"list":    true, // 列表字段总是返回
	"total":   true, // total 字段总是返回（即使为 0）

	// 服务器状态字段
	"online":   true, // online 数组总是返回（即使为空）
	"status":   true, // status 对象总是返回
	"cpu":      true,
	"mem":      true,
	"disk":     true,
	"protocol": true,

	// 用户基础字段
	"telegram":     true,
	"balance":      true,
	"referer_id":   true,
	"user_devices": true,
	"avatar":       true,

	// 用户状态字段
	"enable":              true,
	"is_admin":            true,
	"only_first_purchase": true,
	"enabled":             true,
	"verified":            true,

	// 通知设置字段
	"enable_balance_notify":   true,
	"enable_login_notify":     true,
	"enable_subscribe_notify": true,
	"enable_trade_notify":     true,

	// 营销和优惠字段
	"referral_percentage": true,
	"gift_amount":         true,

	// 迁移和操作状态字段
	"has_migrate": true,
	"success":     true,
	"used_count":  true,
	"user_limit":  true,
	"count":       true,
	"discount":    true,
}

// CustomResponseEncoder 自定义响应编码器
// 必需字段总是输出，其他字段为空时移除
func CustomResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	marshal, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(v.(proto.Message))
	if err != nil {
		return err
	}
	var response []byte
	var data interface{}
	data = make(map[string]interface{})
	if err = json.Unmarshal(marshal, &data); err != nil {
		data = marshal
		response, err = json.Marshal(map[string]interface{}{
			"data":    data,
			"code":    200,
			"message": "success",
		})
	} else {
		if data.(map[string]interface{})["code"] != nil {
			response = marshal
		} else {
			response, err = json.Marshal(map[string]interface{}{
				"data":    data,
				"code":    200,
				"message": "success",
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	return err
}
