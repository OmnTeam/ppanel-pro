package server

import (
	"context"

	commonbiz "github.com/OmnTeam/ppanel-pro/internal/biz/common"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

type compatGetAdsRequest struct {
	Device   string `json:"device"`
	Position string `json:"position"`
}

type compatLegacyAds struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Description string `json:"description"`
	TargetURL   string `json:"target_url"`
	StartTime   int64  `json:"start_time"`
	EndTime     int64  `json:"end_time"`
	Status      int    `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type compatGetAdsResponse struct {
	List []compatLegacyAds `json:"list"`
}

type compatDownloadLink struct {
	IOS     string `json:"ios,omitempty"`
	Android string `json:"android,omitempty"`
	Windows string `json:"windows,omitempty"`
	Mac     string `json:"mac,omitempty"`
	Linux   string `json:"linux,omitempty"`
	Harmony string `json:"harmony,omitempty"`
}

type compatSubscribeClient struct {
	ID           int64              `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description,omitempty"`
	Icon         string             `json:"icon,omitempty"`
	Scheme       string             `json:"scheme,omitempty"`
	IsDefault    bool               `json:"is_default"`
	DownloadLink compatDownloadLink `json:"download_link,omitempty"`
}

type compatGetClientResponse struct {
	Total int64                   `json:"total"`
	List  []compatSubscribeClient `json:"list"`
}

type compatPrivacyPolicyResponse struct {
	PrivacyPolicy string `json:"privacy_policy"`
}

type compatTosResponse struct {
	TosContent string `json:"tos_content"`
}

type compatStatResponse struct {
	User     int64    `json:"user"`
	Node     int64    `json:"node"`
	Country  int64    `json:"country"`
	Protocol []string `json:"protocol"`
}

type compatSendEmailCodeRequest struct {
	Email string `json:"email"`
	Type  int32  `json:"type"`
}

type compatSendSmsCodeRequest struct {
	Type              int32  `json:"type"`
	Telephone         string `json:"telephone"`
	TelephoneAreaCode string `json:"telephone_area_code"`
}

type compatSendCodeResponse struct {
	Code   string `json:"code,omitempty"`
	Status bool   `json:"status"`
}

type compatCheckVerificationCodeRequest struct {
	Method  string `json:"method"`
	Account string `json:"account"`
	Code    string `json:"code"`
	Type    int32  `json:"type"`
}

type compatCheckVerificationCodeResponse struct {
	Status bool `json:"status"`
}

func registerLegacyCommonCompatRoutes(r *khttp.Router, dataLayer *data.Data, appConf *conf.Application, logger log.Logger) {
	if r == nil || dataLayer == nil {
		return
	}

	uc := commonbiz.NewCommonUsecase(data.NewCommonRepo(dataLayer, logger), appConf, logger)

	r.GET("/v1/common/ads", func(ctx khttp.Context) error {
		var req compatGetAdsRequest
		_ = ctx.BindQuery(&req)

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatGetAdsRequest)
			adsList, callErr := uc.GetAds(inner, in.Device, in.Position)
			if callErr != nil {
				return nil, callErr
			}

			resp := &compatGetAdsResponse{List: make([]compatLegacyAds, 0, len(adsList))}
			for _, item := range adsList {
				resp.List = append(resp.List, compatLegacyAds{
					ID:          item.ID,
					Title:       item.Title,
					Type:        item.Type,
					Content:     item.Content,
					Description: item.Description,
					TargetURL:   item.TargetURL,
					StartTime:   item.StartTime,
					EndTime:     item.EndTime,
					Status:      item.Status,
					CreatedAt:   item.CreatedAt,
					UpdatedAt:   item.UpdatedAt,
				})
			}
			return resp, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/common/client", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, request interface{}) (interface{}, error) {
			clientList, total, callErr := uc.GetClient(inner)
			if callErr != nil {
				return nil, callErr
			}

			resp := &compatGetClientResponse{Total: total, List: make([]compatSubscribeClient, 0, len(clientList))}
			for _, item := range clientList {
				resp.List = append(resp.List, compatSubscribeClient{
					ID:          item.ID,
					Name:        item.Name,
					Description: item.Description,
					Icon:        item.Icon,
					Scheme:      item.Scheme,
					IsDefault:   item.IsDefault,
					DownloadLink: compatDownloadLink{
						IOS:     item.DownloadLink.IOS,
						Android: item.DownloadLink.Android,
						Windows: item.DownloadLink.Windows,
						Mac:     item.DownloadLink.Mac,
						Linux:   item.DownloadLink.Linux,
						Harmony: item.DownloadLink.Harmony,
					},
				})
			}
			return resp, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/common/site/privacy", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, request interface{}) (interface{}, error) {
			content, callErr := uc.GetPrivacyPolicy(inner)
			if callErr != nil {
				return nil, callErr
			}
			return &compatPrivacyPolicyResponse{PrivacyPolicy: content}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/common/site/tos", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, request interface{}) (interface{}, error) {
			content, callErr := uc.GetTos(inner)
			if callErr != nil {
				return nil, callErr
			}
			return &compatTosResponse{TosContent: content}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.GET("/v1/common/site/stat", func(ctx khttp.Context) error {
		out, err := compatMiddleware(ctx, nil, func(inner context.Context, request interface{}) (interface{}, error) {
			stat, callErr := uc.GetStat(inner)
			if callErr != nil {
				return nil, callErr
			}
			return &compatStatResponse{
				User:     stat.User,
				Node:     stat.Node,
				Country:  stat.Country,
				Protocol: stat.Protocol,
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.POST("/v1/common/send_code", func(ctx khttp.Context) error {
		var req compatSendEmailCodeRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatSendEmailCodeRequest)
			code, callErr := uc.SendEmailCode(inner, in.Email, in.Type)
			if callErr != nil {
				return nil, callErr
			}
			return &compatSendCodeResponse{
				Code:   compatMaybeExposeVerifyCode(code),
				Status: true,
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.POST("/v1/common/send_sms_code", func(ctx khttp.Context) error {
		var req compatSendSmsCodeRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatSendSmsCodeRequest)
			code, callErr := uc.SendSmsCode(inner, in.Telephone, in.TelephoneAreaCode, in.Type)
			if callErr != nil {
				return nil, callErr
			}
			return &compatSendCodeResponse{
				Code:   compatMaybeExposeVerifyCode(code),
				Status: true,
			}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})

	r.POST("/v1/common/check_verification_code", func(ctx khttp.Context) error {
		var req compatCheckVerificationCodeRequest
		if err := ctx.Bind(&req); err != nil {
			return compatJSONError(ctx, err)
		}
		if err := ctx.BindQuery(&req); err != nil {
			return compatJSONError(ctx, err)
		}

		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			in := request.(*compatCheckVerificationCodeRequest)
			valid, callErr := uc.CheckVerificationCode(inner, in.Method, in.Account, in.Code, in.Type)
			if callErr != nil {
				return nil, callErr
			}
			return &compatCheckVerificationCodeResponse{Status: valid}, nil
		})
		if err != nil {
			return compatJSONError(ctx, err)
		}
		return compatJSON(ctx, out)
	})
}

func compatMaybeExposeVerifyCode(code string) string {
	_ = code
	return ""
}
