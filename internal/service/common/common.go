package common

import (
	"context"
	"strconv"

	pb "github.com/OmnTeam/ppanel-pro/api/public/common/v1"
	"github.com/OmnTeam/ppanel-pro/internal/biz/common"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

type CommonService struct {
	pb.UnimplementedCommonServer

	uc *common.CommonUsecase
}

func NewCommonService(uc *common.CommonUsecase) *CommonService {
	return &CommonService{
		uc: uc,
	}
}

// GetAds gets ads list
func (s *CommonService) GetAds(ctx context.Context, req *pb.GetAdsRequest) (*pb.GetAdsReply, error) {
	adsList, err := s.uc.GetAds(ctx, req.Device, req.Position)
	if err != nil {
		return nil, err
	}

	// Convert biz objects to proto objects
	pbAds := make([]*pb.Ads, len(adsList))
	for i, ad := range adsList {
		pbAds[i] = &pb.Ads{
			Id:          ad.ID,
			Title:       ad.Title,
			Type:        ad.Type,
			Content:     ad.Content,
			Description: ad.Description,
			TargetUrl:   ad.TargetURL,
			StartTime:   ad.StartTime,
			EndTime:     ad.EndTime,
			Status:      int32(ad.Status),
			CreatedAt:   ad.CreatedAt,
			UpdatedAt:   ad.UpdatedAt,
		}
	}

	return &pb.GetAdsReply{
		Code:    int32(responsecode.GetAdsSuccess),
		Message: responsecode.CodeMessages[responsecode.GetAdsSuccess],
		Data: &pb.GetAdsData{
			List: pbAds,
		},
	}, nil
}

// GetClient gets subscribe client list
func (s *CommonService) GetClient(ctx context.Context, req *pb.GetClientRequest) (*pb.GetClientReply, error) {
	clientList, total, err := s.uc.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	// Convert biz objects to proto objects
	pbClients := make([]*pb.SubscribeClient, len(clientList))
	for i, client := range clientList {
		pbClients[i] = &pb.SubscribeClient{
			Id:          client.ID,
			Name:        client.Name,
			Description: client.Description,
			Icon:        client.Icon,
			Scheme:      client.Scheme,
			IsDefault:   client.IsDefault,
			DownloadLink: &pb.DownloadLink{
				Ios:     client.DownloadLink.IOS,
				Android: client.DownloadLink.Android,
				Windows: client.DownloadLink.Windows,
				Mac:     client.DownloadLink.Mac,
				Linux:   client.DownloadLink.Linux,
				Harmony: client.DownloadLink.Harmony,
			},
		}
	}

	return &pb.GetClientReply{
		Code:    int32(responsecode.GetClientSuccess),
		Message: responsecode.CodeMessages[responsecode.GetClientSuccess],
		Data: &pb.GetClientData{
			Total: total,
			List:  pbClients,
		},
	}, nil
}

// GetPrivacyPolicy gets privacy policy content
func (s *CommonService) GetPrivacyPolicy(ctx context.Context, req *pb.GetPrivacyPolicyRequest) (*pb.GetPrivacyPolicyReply, error) {
	content, err := s.uc.GetPrivacyPolicy(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetPrivacyPolicyReply{
		Code:    int32(responsecode.GetPrivacyPolicySuccess),
		Message: responsecode.CodeMessages[responsecode.GetPrivacyPolicySuccess],
		Data: &pb.GetPrivacyPolicyData{
			PrivacyPolicy: content,
		},
	}, nil
}

// GetTos gets terms of service content
func (s *CommonService) GetTos(ctx context.Context, req *pb.GetTosRequest) (*pb.GetTosReply, error) {
	content, err := s.uc.GetTos(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetTosReply{
		Code:    int32(responsecode.GetTosSuccess),
		Message: responsecode.CodeMessages[responsecode.GetTosSuccess],
		Data: &pb.GetTosData{
			TosContent: content,
		},
	}, nil
}

// GetGlobalConfig gets global configuration
func (s *CommonService) GetGlobalConfig(ctx context.Context, req *pb.GetGlobalConfigRequest) (*pb.GetGlobalConfigReply, error) {
	config, err := s.uc.GetGlobalConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Convert config to proto structures
	configData := &pb.GetGlobalConfigData{
		Site: &pb.SiteConfig{
			Host:       config.Site.Host,
			SiteName:   config.Site.SiteName,
			SiteDesc:   config.Site.SiteDesc,
			SiteLogo:   config.Site.SiteLogo,
			Keywords:   config.Site.Keywords,
			CustomHtml: config.Site.CustomHtml,
			CustomData: config.Site.CustomData,
		},
		Verify: &pb.VerifyConfig{
			TurnstileSiteKey:          config.Verify.TurnstileSiteKey,
			EnableLoginVerify:         config.Verify.EnableLoginVerify,
			EnableRegisterVerify:      config.Verify.EnableRegisterVerify,
			EnableResetPasswordVerify: config.Verify.EnableResetPasswordVerify,
		},
		Auth: &pb.AuthConfig{
			Mobile: &pb.MobileAuthConfig{
				Enable:          config.Auth.Mobile.Enable,
				EnableWhitelist: config.Auth.Mobile.EnableWhitelist,
				Whitelist:       config.Auth.Mobile.Whitelist,
			},
			Email: &pb.EmailAuthConfig{
				Enable:             config.Auth.Email.Enable,
				EnableVerify:       config.Auth.Email.EnableVerify,
				EnableDomainSuffix: config.Auth.Email.EnableDomainSuffix,
				DomainSuffixList:   config.Auth.Email.DomainSuffixList,
			},
			Register: &pb.RegisterConfig{
				StopRegister:            config.Auth.Register.StopRegister,
				EnableIpRegisterLimit:   config.Auth.Register.EnableIpRegisterLimit,
				IpRegisterLimit:         config.Auth.Register.IpRegisterLimit,
				IpRegisterLimitDuration: config.Auth.Register.IpRegisterLimitDuration,
			},
		},
		Invite: &pb.InviteConfig{
			ForcedInvite:       config.Invite.ForcedInvite,
			ReferralPercentage: config.Invite.ReferralPercentage,
			OnlyFirstPurchase:  config.Invite.OnlyFirstPurchase,
		},
		Currency: &pb.CurrencyConfig{
			CurrencyUnit:   config.Currency["currency_unit"],
			CurrencySymbol: config.Currency["currency_symbol"],
		},
		Subscribe: &pb.SubscribeConfig{
			SingleModel:     config.Subscribe.SingleModel,
			SubscribePath:   config.Subscribe.SubscribePath,
			SubscribeDomain: config.Subscribe.SubscribeDomain,
			PanDomain:       config.Subscribe.PanDomain,
			UserAgentLimit:  config.Subscribe.UserAgentLimit,
			UserAgentList:   config.Subscribe.UserAgentList,
		},
		VerifyCode: &pb.PublicVerifyCodeConfig{
			// Convert verify_code_interval from string to int64
			VerifyCodeInterval: 60, // Default 60 seconds
		},
		OauthMethods: config.OAuthMethods,
		WebAd:        config.WebAd,
	}

	// Try to parse verify_code_interval from database config
	if interval, ok := config.VerifyCode["verify_code_interval"]; ok {
		if val, err := strconv.ParseInt(interval, 10, 64); err == nil {
			configData.VerifyCode.VerifyCodeInterval = val
		}
	}

	return &pb.GetGlobalConfigReply{
		Code:    int32(responsecode.GetGlobalConfigSuccess),
		Message: responsecode.CodeMessages[responsecode.GetGlobalConfigSuccess],
		Data:    configData,
	}, nil
}

// GetStat gets system statistics
func (s *CommonService) GetStat(ctx context.Context, req *pb.GetStatRequest) (*pb.GetStatReply, error) {
	stat, err := s.uc.GetStat(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetStatReply{
		Code:    int32(responsecode.GetStatSuccess),
		Message: responsecode.CodeMessages[responsecode.GetStatSuccess],
		Data: &pb.GetStatData{
			User:     stat.User,
			Node:     stat.Node,
			Country:  stat.Country,
			Protocol: stat.Protocol,
		},
	}, nil
}

// SendEmailCode sends email verification code
func (s *CommonService) SendEmailCode(ctx context.Context, req *pb.SendEmailCodeRequest) (*pb.SendCodeReply, error) {
	code, err := s.uc.SendEmailCode(ctx, req.Email, req.Type)
	if err != nil {
		return nil, err
	}

	return &pb.SendCodeReply{
		Code:    int32(responsecode.SendEmailCodeSuccess),
		Message: responsecode.CodeMessages[responsecode.SendEmailCodeSuccess],
		Data: &pb.SendCodeData{
			Status: true,
			Code:   code, // Only returned in development mode
		},
	}, nil
}

// SendSmsCode sends SMS verification code
func (s *CommonService) SendSmsCode(ctx context.Context, req *pb.SendSmsCodeRequest) (*pb.SendCodeReply, error) {
	code, err := s.uc.SendSmsCode(ctx, req.Telephone, req.TelephoneAreaCode, req.Type)
	if err != nil {
		return nil, err
	}

	return &pb.SendCodeReply{
		Code:    int32(responsecode.SendSmsCodeSuccess),
		Message: responsecode.CodeMessages[responsecode.SendSmsCodeSuccess],
		Data: &pb.SendCodeData{
			Status: true,
			Code:   code, // Only returned in development mode
		},
	}, nil
}

// CheckVerificationCode checks verification code
func (s *CommonService) CheckVerificationCode(ctx context.Context, req *pb.CheckVerificationCodeRequest) (*pb.CheckVerificationCodeReply, error) {
	valid, err := s.uc.CheckVerificationCode(ctx, req.Method, req.Account, req.Code, req.Type)
	if err != nil {
		return nil, err
	}

	return &pb.CheckVerificationCodeReply{
		Code:    int32(responsecode.CheckVerificationCodeSuccess),
		Message: responsecode.CodeMessages[responsecode.CheckVerificationCodeSuccess],
		Data: &pb.CheckVerificationCodeData{
			Status: valid,
		},
	}, nil
}
