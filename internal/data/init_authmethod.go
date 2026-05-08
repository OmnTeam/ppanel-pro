package data

import (
	"context"
	"encoding/json"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/npanel-pro/internal/model/auth"
	"github.com/go-kratos/kratos/v2/log"
)

// initAuthMethodData 初始化认证方法配置数据
func initAuthMethodData(client *ent.Client, logger log.Logger) error {
	ctx := context.Background()
	helper := log.NewHelper(logger)

	// 检查 mobile 配置是否已存在
	exists, err := client.ProxyAuthMethod.
		Query().
		Where(
			proxyauthmethod.Method("mobile"),
		).
		Exist(ctx)

	if err != nil {
		helper.Errorf("检查 mobile 配置失败: %v", err)
		return err
	}

	if exists {
		helper.Info("mobile 配置已存在，跳过初始化")
		return nil
	}

	// 创建 abosend 平台配置
	abosendConfig := auth.AbosendConfig{
		ApiDomain: "https://smsapi.abosend.com",
		Access:    "UVTtbbTz",
		Secret:    "CVRZQVJLTJWTBDXDWSYSOITEWLUMBRCO",
		Template:  "Your verification code is: {{.code}}",
	}

	// 创建 mobile 认证配置
	mobileConfig := auth.MobileAuthConfig{
		Platform:        "abosend",
		PlatformConfig:  abosendConfig,
		EnableWhitelist: false,
		Whitelist:       []string{},
	}

	// 序列化配置
	configJSON := mobileConfig.Marshal()

	// 插入到数据库
	_, err = client.ProxyAuthMethod.
		Create().
		SetMethod("mobile").
		SetConfig(configJSON).
		SetEnabled(true).
		Save(ctx)

	if err != nil {
		helper.Errorf("插入 mobile 配置失败: %v", err)
		return err
	}

	helper.Info("成功初始化 mobile 认证配置（abosend 平台）")

	// 打印配置用于验证
	var prettyConfig map[string]interface{}
	json.Unmarshal([]byte(configJSON), &prettyConfig)
	prettyJSON, _ := json.MarshalIndent(prettyConfig, "", "  ")
	helper.Infof("配置内容:\n%s", string(prettyJSON))

	return nil
}
