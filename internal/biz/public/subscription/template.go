package subscription

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

// TemplateData 模板数据结构
type TemplateData struct {
	SiteName      string
	SubscribeName string
	OutputFormat  string
	Proxies       []map[string]interface{}
	UserInfo      UserInfo
	Params        map[string]string
}

// RenderTemplate 渲染订阅配置模板（按照原项目逻辑）
func RenderTemplate(
	templateStr string,
	outputFormat string,
	siteName string,
	subscribeName string,
	nodes []*NodeInfo,
	userSubscribe *UserSubscribe,
	userInfo UserInfo,
	params map[string]string,
) ([]byte, error) {
	// 1. 转换节点为Proxy格式
	proxies := make([]map[string]interface{}, 0, len(nodes))
	for _, node := range nodes {
		proxyMap := structToMap(node)
		proxies = append(proxies, proxyMap)
	}

	// 2. 构建用户信息（使用传入的userInfo，其中包含订阅URL）
	if userInfo.Password == "" {
		userInfo.Password = userSubscribe.UUID
	}
	if userInfo.ExpiredAt.IsZero() && userSubscribe.ExpireTime > 0 {
		userInfo.ExpiredAt = time.UnixMilli(userSubscribe.ExpireTime)
	}

	// 3. 构建模板数据
	templateData := TemplateData{
		SiteName:      siteName,
		SubscribeName: subscribeName,
		OutputFormat:  outputFormat,
		Proxies:       proxies,
		UserInfo:      userInfo,
		Params:        params,
	}

	// 4. 解析模板
	tmpl, err := template.New("subscribe").Funcs(templateFuncMap()).Parse(templateStr)
	if err != nil {
		return nil, err
	}

	// 5. 执行模板渲染
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		return nil, err
	}

	result := buf.String()

	// 6. 根据输出格式处理
	if outputFormat == "base64" {
		encoded := base64.StdEncoding.EncodeToString([]byte(result))
		return []byte(encoded), nil
	}

	return buf.Bytes(), nil
}

func templateFuncMap() template.FuncMap {
	funcs := sprig.TxtFuncMap()
	funcs["simnetHexPSK"] = simnetHexPSK
	funcs["buildOmnxtSimnetConfigs"] = buildOmnxtSimnetConfigs
	return funcs
}

func simnetHexPSK(psk string) string {
	trimmed := strings.TrimSpace(psk)
	if trimmed == "" {
		return ""
	}
	if len(trimmed)%2 == 0 {
		if _, err := hex.DecodeString(trimmed); err == nil {
			return strings.ToLower(trimmed)
		}
	}
	return hex.EncodeToString([]byte(trimmed))
}

func buildOmnxtSimnetConfigs(proxies []map[string]interface{}, params map[string]string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	proxyMode := strings.TrimSpace(params["proxy_mode"])
	if proxyMode == "" {
		proxyMode = "global"
	}

	dnsServers := []string{"1.1.1.1"}
	if raw := strings.TrimSpace(params["dns_servers"]); raw != "" {
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == '\n' || r == '\r'
		})
		parsed := make([]string, 0, len(parts))
		for _, item := range parts {
			item = strings.TrimSpace(item)
			if item != "" {
				parsed = append(parsed, item)
			}
		}
		if len(parsed) > 0 {
			dnsServers = parsed
		}
	}

	for _, proxy := range proxies {
		if mapString(proxy["Type"]) != "simnet" {
			continue
		}

		item := map[string]interface{}{
			"tag":                          mapString(proxy["Name"]),
			"server_addr":                  mapString(proxy["Server"]),
			"server_port":                  mapInt(proxy["Port"]),
			"protocol":                     "simnet",
			"sni":                          mapString(proxy["SNI"]),
			"allow_insecure":               mapBool(proxy["AllowInsecure"]),
			"simnet_psk":                   simnetHexPSK(mapString(proxy["SimnetPsk"])),
			"simnet_key_id":                mapInt(proxy["SimnetKeyID"]),
			"simnet_ticket_id":             mapStringOrNil(proxy["SimnetTicketID"]),
			"simnet_path":                  mapStringOrNil(proxy["SimnetPath"]),
			"simnet_carrier":               defaultString(mapString(proxy["SimnetCarrier"]), "h2"),
			"simnet_af_enabled":            mapBool(proxy["SimnetAfEnabled"]),
			"simnet_af_path_mode":          defaultString(mapString(proxy["SimnetAfPathMode"]), "api"),
			"simnet_af_path_prefix":        mapStringOrNil(proxy["SimnetAfPathPrefix"]),
			"simnet_af_path_suffix":        mapStringOrNil(proxy["SimnetAfPathSuffix"]),
			"simnet_af_magic_mode":         defaultString(mapString(proxy["SimnetAfMagicMode"]), "derived"),
			"simnet_af_response_jitter_ms": defaultInt(mapInt(proxy["SimnetAfResponseJitterMs"]), 50),
			"proxy_mode":                   proxyMode,
			"dns_servers":                  dnsServers,
		}
		result = append(result, item)
	}

	return result
}

func mapString(value interface{}) string {
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func mapBool(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

func mapInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int8:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case int64:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case uint64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		n, _ := v.Int64()
		return int(n)
	default:
		return 0
	}
}

func mapStringOrNil(value interface{}) interface{} {
	if s := mapString(value); s != "" {
		return s
	}
	return nil
}

func defaultString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func defaultInt(value, fallback int) int {
	if value != 0 {
		return value
	}
	return fallback
}

// structToMap 将结构体转换为map（按照原项目逻辑）
func structToMap(obj interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	v := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			m[field.Name] = v.Field(i).Interface()
		}
	}

	return m
}
