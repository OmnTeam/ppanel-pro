package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-kratos/kratos/v2/errors"
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

// CustomRequestDecoder 自定义请求解码器
// 处理前端发送的 {"enabled": {}} 空对象，转换为 null
// 支持 protobuf json_name，允许接收 camelCase 字段名（如 orderNo）
func CustomRequestDecoder(r *http.Request, v interface{}) error {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// 如果请求体为空，直接返回
	if len(body) == 0 {
		return nil
	}

	// 预处理：将空对象 {} 替换为 null
	// 例如：{"enabled": {}} -> {"enabled": null}
	processedBody := preprocessEmptyObjects(body)

	// 预处理：将 protobuf json_name (camelCase) 转换为 JSON tag (snake_case)
	// 例如：{"orderNo": "xxx"} -> {"order_no": "xxx"}
	processedBody = convertProtoJSONNames(processedBody, v)

	// 使用标准 JSON 解码
	decoder := json.NewDecoder(bytes.NewReader(processedBody))
	decoder.UseNumber() // 保持数字精度
	if err := decoder.Decode(v); err != nil {
		return err
	}

	return nil
}

// preprocessEmptyObjects 预处理 JSON，将空对象 {} 替换为 null
func preprocessEmptyObjects(data []byte) []byte {
	// 解析为 map
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		// 如果解析失败，返回原始数据
		return data
	}

	// 递归处理
	cleanEmptyObjects(rawData)

	// 重新序列化
	result, err := json.Marshal(rawData)
	if err != nil {
		return data
	}
	return result
}

// cleanEmptyObjects 递归清理空对象，将其替换为 nil
func cleanEmptyObjects(data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if emptyMap, ok := value.(map[string]interface{}); ok && len(emptyMap) == 0 {
				// 空对象替换为 nil
				v[key] = nil
			} else {
				// 递归处理
				cleanEmptyObjects(value)
			}
		}
	case []interface{}:
		for _, item := range v {
			cleanEmptyObjects(item)
		}
	}
}

// convertProtoJSONNames 将请求 JSON 中的 protobuf json_name 转换为 Go struct JSON tag
// 例如：{"orderNo": "xxx"} -> {"order_no": "xxx"}
func convertProtoJSONNames(data []byte, v interface{}) []byte {
	// 解析为 map
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		// 如果解析失败，返回原始数据
		return data
	}

	// 构建 protobuf json_name 到 JSON tag 的映射
	mapping := buildProtoJSONMapping(reflect.TypeOf(v))

	// 转换字段名
	convertedData := convertMapKeys(rawData, mapping)

	// 重新序列化
	result, err := json.Marshal(convertedData)
	if err != nil {
		return data
	}
	return result
}

// buildProtoJSONMapping 构建 protobuf json_name 到 JSON tag 的映射
func buildProtoJSONMapping(t reflect.Type) map[string]string {
	mapping := make(map[string]string)

	// 处理指针类型
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return mapping
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取 JSON 标签名
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// 解析 JSON 标签（移除 omitempty 等选项）
		jsonTagName := jsonTag
		for idx := 0; idx < len(jsonTag); idx++ {
			if jsonTag[idx] == ',' {
				jsonTagName = jsonTag[:idx]
				break
			}
		}

		// 获取 protobuf 标签中的 json_name
		if protoTag := field.Tag.Get("protobuf"); protoTag != "" {
			// 查找 json= 部分
			if jsonStart := findSubstring(protoTag, "json="); jsonStart != -1 {
				jsonStart += 5 // 跳过 "json="
				jsonEnd := jsonStart
				for jsonEnd < len(protoTag) && protoTag[jsonEnd] != ',' {
					jsonEnd++
				}
				if jsonEnd > jsonStart {
					protoJSONName := protoTag[jsonStart:jsonEnd]
					// 建立映射：protobuf json_name -> Go JSON tag
					// 例如：orderNo -> order_no
					mapping[protoJSONName] = jsonTagName
				}
			}
		}
	}

	return mapping
}

// convertMapKeys 递归转换 map 的键名
func convertMapKeys(data map[string]interface{}, mapping map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// 查找是否需要转换键名
		newKey := key
		if mappedKey, ok := mapping[key]; ok {
			newKey = mappedKey
		}

		// 递归处理嵌套对象
		switch v := value.(type) {
		case map[string]interface{}:
			result[newKey] = convertMapKeys(v, mapping)
		case []interface{}:
			result[newKey] = convertArrayValues(v, mapping)
		default:
			result[newKey] = value
		}
	}

	return result
}

// convertArrayValues 递归转换数组中的值
func convertArrayValues(data []interface{}, mapping map[string]string) []interface{} {
	result := make([]interface{}, len(data))

	for i, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			result[i] = convertMapKeys(v, mapping)
		case []interface{}:
			result[i] = convertArrayValues(v, mapping)
		default:
			result[i] = item
		}
	}

	return result
}

// CustomResponseEncoder 自定义响应编码器
// 必需字段总是输出，其他字段为空时移除
func CustomResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	// 自定义序列化，必需字段保留零值，其他字段移除零值
	data, err := marshalWithRequiredFields(v)
	if err != nil {
		return err
	}
	jsonData := make(map[string]interface{})
	err = json.Unmarshal(data, &jsonData)
	if err == nil {
		sprintf := fmt.Sprintf("%v", jsonData["code"])
		if len(sprintf) > 1 && sprintf[0] == '2' {
			jsonData["code"] = 200
			data, err = json.Marshal(jsonData)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

// marshalWithRequiredFields 序列化，保留必需字段的零值
func marshalWithRequiredFields(v interface{}) ([]byte, error) {
	// 转换为 map 以便处理字段
	result := toMapWithRequiredFields(reflect.ValueOf(v))
	return json.Marshal(result)
}

// toMapWithRequiredFields 将结构体转为 map，保留必需字段的零值
func toMapWithRequiredFields(v reflect.Value) interface{} {
	// 处理指针
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		result := make(map[string]interface{})
		t := v.Type()

		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)

			// 跳过未导出的字段
			if !field.IsExported() {
				continue
			}

			// 获取 JSON 标签名
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// 解析 JSON 标签（移除 omitempty 等选项）
			tagName := jsonTag
			for idx := 0; idx < len(jsonTag); idx++ {
				if jsonTag[idx] == ',' {
					tagName = jsonTag[:idx]
					break
				}
			}

			// 检查 protobuf 标签中的 json_name（用于与老项目兼容）
			// 只有当 protobuf 的 json_name 与字段名的驼峰格式不同时，才使用（说明是手动指定的）
			// 例如: protobuf:"bytes,9,opt,name=subscribe_template,json=template,proto3"
			if protoTag := field.Tag.Get("protobuf"); protoTag != "" {
				// 查找 json= 部分
				if jsonStart := findSubstring(protoTag, "json="); jsonStart != -1 {
					jsonStart += 5 // 跳过 "json="
					jsonEnd := jsonStart
					for jsonEnd < len(protoTag) && protoTag[jsonEnd] != ',' {
						jsonEnd++
					}
					if jsonEnd > jsonStart {
						protoJSONName := protoTag[jsonStart:jsonEnd]
						// 只有当 protobuf json_name 与字段名驼峰格式不同时才使用
						// 例如: SubscribeTemplate -> subscribeTemplate (自动生成，不使用)
						//       SubscribeTemplate -> template (手动指定，使用)
						fieldNameCamel := toCamelCase(field.Name)
						if protoJSONName != fieldNameCamel {
							// 使用手动指定的 protobuf json_name
							tagName = protoJSONName
						}
					}
				}
			}

			// 递归处理字段值
			fieldResult := toMapWithRequiredFields(fieldValue)

			// 判断是否保留该字段
			if requiredFields[tagName] {
				// 必需字段总是保留
				result[tagName] = fieldResult
			} else if !isZeroValue(fieldValue) {
				// 非必需字段只保留非零值
				result[tagName] = fieldResult
			}
		}
		return result

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return []interface{}{}
		}
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = toMapWithRequiredFields(v.Index(i))
		}
		return result

	case reflect.Map:
		if v.Len() == 0 {
			return map[string]interface{}{}
		}
		result := make(map[string]interface{})
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			result[key] = toMapWithRequiredFields(iter.Value())
		}
		return result

	default:
		// 基本类型直接返回
		return v.Interface()
	}
}

// toCamelCase 将字段名转换为驼峰格式（首字母小写）
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	// 首字母小写（只对大写字母转换）
	runes := []rune(s)
	if runes[0] >= 'A' && runes[0] <= 'Z' {
		runes[0] = runes[0] + ('a' - 'A')
	}
	return string(runes)
}

// findSubstring 查找子字符串的位置（类似 strings.Index，但不引入额外依赖）
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// isZeroValue 判断是否为零值
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		// 结构体需要检查所有字段
		return false
	default:
		return false
	}
}
