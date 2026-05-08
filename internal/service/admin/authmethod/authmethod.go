package authmethod

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/structpb"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/authmethod/v1"
	authmethodbiz "github.com/OmnTeam/npanel-pro/internal/biz/admin/authmethod"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
)

// AuthMethodService 认证方法服务
type AuthMethodService struct {
	v1.UnimplementedAuthMethodServiceServer

	uc     *authmethodbiz.AuthMethodUsecase
	logger *log.Helper
}

// NewAuthMethodService 创建认证方法服务
func NewAuthMethodService(uc *authmethodbiz.AuthMethodUsecase, logger log.Logger) *AuthMethodService {
	return &AuthMethodService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// GetAuthMethodConfig 获取认证方法配置
func (s *AuthMethodService) GetAuthMethodConfig(ctx context.Context, req *v1.GetAuthMethodConfigRequest) (*v1.AuthMethodConfigReply, error) {
	auth, err := s.uc.GetAuthMethodConfig(ctx, req.Method)
	if err != nil {
		return nil, err
	}

	config, _ := s.parseConfig(auth.Config)
	return &v1.AuthMethodConfigReply{
		Code:    int32(responsecode.AdminGetAuthMethodConfigSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetAuthMethodConfigSuccess],
		Data: &v1.AuthMethodConfigData{
			Config: &v1.AuthMethodConfig{
				Id:      auth.ID,
				Method:  auth.Method,
				Config:  config,
				Enabled: auth.Enabled,
			},
		},
	}, nil
}

// UpdateAuthMethodConfig 更新认证方法配置
func (s *AuthMethodService) UpdateAuthMethodConfig(ctx context.Context, req *v1.UpdateAuthMethodConfigRequest) (*v1.AuthMethodConfigReply, error) {
	// 构造 biz 层请求（对应原项目的 types.UpdateAuthMethodConfigRequest）
	bizReq := &authmethodbiz.UpdateAuthMethodRequest{
		ID:     req.Id,
		Method: req.Method,
	}
	if req.Enabled != nil {
		bizReq.Enabled = &req.Enabled.Value
	}

	// 将 protobuf Struct 转换为 interface{}（对应原项目的 req.Config）
	if req.Config != nil {
		bizReq.Config = req.Config.AsMap()
	}

	// 调用 biz 层更新配置
	result, err := s.uc.UpdateAuthMethodConfig(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	// 对应原项目第75-82行：构造响应并解析 Config
	config, _ := s.parseConfig(result.Config)
	return &v1.AuthMethodConfigReply{
		Code:    int32(responsecode.AdminUpdateAuthMethodConfigSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateAuthMethodConfigSuccess],
		Data: &v1.AuthMethodConfigData{
			Config: &v1.AuthMethodConfig{
				Id:      result.ID,
				Method:  result.Method,
				Config:  config,
				Enabled: result.Enabled,
			},
		},
	}, nil
}

// GetEmailPlatform 获取邮件平台列表
func (s *AuthMethodService) GetEmailPlatform(ctx context.Context, req *v1.GetEmailPlatformRequest) (*v1.PlatformListReply, error) {
	// 平台列表是全局配置
	platforms := s.uc.GetEmailPlatforms(ctx)
	result := make([]*v1.Platform, 0, len(platforms))
	for _, p := range platforms {
		result = append(result, &v1.Platform{
			Platform:                 p.Platform,
			PlatformUrl:              p.PlatformUrl,
			PlatformFieldDescription: p.PlatformFieldDescription,
		})
	}
	return &v1.PlatformListReply{
		Code:    int32(responsecode.AdminGetEmailPlatformSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetEmailPlatformSuccess],
		Data: &v1.PlatformListData{
			Platforms: result,
		},
	}, nil
}

// GetSmsPlatform 获取短信平台列表
func (s *AuthMethodService) GetSmsPlatform(ctx context.Context, req *v1.GetSmsPlatformRequest) (*v1.PlatformListReply, error) {
	// 平台列表是全局配置
	platforms := s.uc.GetSmsPlatforms(ctx)
	result := make([]*v1.Platform, 0, len(platforms))
	for _, p := range platforms {
		result = append(result, &v1.Platform{
			Platform:                 p.Platform,
			PlatformUrl:              p.PlatformUrl,
			PlatformFieldDescription: p.PlatformFieldDescription,
		})
	}
	return &v1.PlatformListReply{
		Code:    int32(responsecode.AdminGetSmsPlatformSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetSmsPlatformSuccess],
		Data: &v1.PlatformListData{
			Platforms: result,
		},
	}, nil
}

// GetAuthMethodList 获取认证方法列表
func (s *AuthMethodService) GetAuthMethodList(ctx context.Context, req *v1.GetAuthMethodListRequest) (*v1.AuthMethodListReply, error) {
	list, err := s.uc.GetAuthMethodList(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*v1.AuthMethodConfig, 0, len(list))
	for _, auth := range list {
		config, _ := s.parseConfig(auth.Config)
		result = append(result, &v1.AuthMethodConfig{
			Id:      auth.ID,
			Method:  auth.Method,
			Config:  config,
			Enabled: auth.Enabled,
		})
	}
	return &v1.AuthMethodListReply{
		Code:    int32(responsecode.AdminGetAuthMethodListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetAuthMethodListSuccess],
		Data: &v1.AuthMethodListData{
			List: result,
		},
	}, nil
}

// TestEmailSend 测试邮件发送
func (s *AuthMethodService) TestEmailSend(ctx context.Context, req *v1.TestEmailSendRequest) (*v1.TestSendReply, error) {
	success, message, err := s.uc.TestEmailSend(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	return &v1.TestSendReply{
		Code:    int32(responsecode.AdminTestEmailSendSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminTestEmailSendSuccess],
		Data: &v1.TestSendData{
			Success:       success,
			ResultMessage: message,
		},
	}, nil
}

// TestSmsSend 测试短信发送
func (s *AuthMethodService) TestSmsSend(ctx context.Context, req *v1.TestSmsSendRequest) (*v1.TestSendReply, error) {
	success, message, err := s.uc.TestSmsSendWithAreaCode(ctx, req.AreaCode, req.Telephone)
	if err != nil {
		return nil, err
	}
	return &v1.TestSendReply{
		Code:    int32(responsecode.AdminTestSmsSendSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminTestSmsSendSuccess],
		Data: &v1.TestSendData{
			Success:       success,
			ResultMessage: message,
		},
	}, nil
}

// parseConfig 解析JSON配置为Struct
func (s *AuthMethodService) parseConfig(configJSON string) (*structpb.Struct, error) {
	if configJSON == "" {
		return &structpb.Struct{}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &data); err != nil {
		return &structpb.Struct{}, err
	}

	return structpb.NewStruct(data)
}

// marshalConfig 将Struct转换为JSON
func (s *AuthMethodService) marshalConfig(config *structpb.Struct) (string, error) {
	if config == nil {
		return "{}", nil
	}

	data, err := json.Marshal(config.AsMap())
	if err != nil {
		return "{}", err
	}

	return string(data), nil
}
