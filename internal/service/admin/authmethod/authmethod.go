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

func NewAuthMethodService(uc *authmethodbiz.AuthMethodUsecase, logger log.Logger) *AuthMethodService {
	return &AuthMethodService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

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

func (s *AuthMethodService) UpdateAuthMethodConfig(ctx context.Context, req *v1.UpdateAuthMethodConfigRequest) (*v1.AuthMethodConfigReply, error) {
	bizReq := &authmethodbiz.UpdateAuthMethodRequest{
		ID:     req.Id,
		Method: req.Method,
	}
	if req.Enabled != nil {
		bizReq.Enabled = &req.Enabled.Value
	}
	if req.Config != nil {
		bizReq.Config = req.Config.AsMap()
	}

	result, err := s.uc.UpdateAuthMethodConfig(ctx, bizReq)
	if err != nil {
		return nil, err
	}

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

func (s *AuthMethodService) GetEmailPlatform(ctx context.Context, req *v1.GetEmailPlatformRequest) (*v1.PlatformListReply, error) {
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
		Code: 200,
		Msg:  "success",
		Data: &v1.PlatformListData{
			List: result,
		},
	}, nil
}

func (s *AuthMethodService) GetSmsPlatform(ctx context.Context, req *v1.GetSmsPlatformRequest) (*v1.PlatformListReply, error) {
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
		Code: 200,
		Msg:  "success",
		Data: &v1.PlatformListData{
			List: result,
		},
	}, nil
}

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

func (s *AuthMethodService) TestEmailSend(ctx context.Context, req *v1.TestEmailSendRequest) (*v1.ActionReply, error) {
	_, _, err := s.uc.TestEmailSend(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	return &v1.ActionReply{
		Code: 200,
		Msg:  "success",
	}, nil
}

func (s *AuthMethodService) TestSmsSend(ctx context.Context, req *v1.TestSmsSendRequest) (*v1.ActionReply, error) {
	_, _, err := s.uc.TestSmsSendWithAreaCode(ctx, req.AreaCode, req.Telephone)
	if err != nil {
		return nil, err
	}
	return &v1.ActionReply{
		Code: 200,
		Msg:  "success",
	}, nil
}

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
