package subscription

import (
	"bytes"
	"encoding/base64"
	"reflect"
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
	tmpl, err := template.New("subscribe").Funcs(sprig.TxtFuncMap()).Parse(templateStr)
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
