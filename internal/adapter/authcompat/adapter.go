package authcompat

import (
	"context"

	authbiz "github.com/OmnTeam/npanel-pro/internal/biz/auth"
	"github.com/OmnTeam/npanel-pro/internal/data"
	authservice "github.com/OmnTeam/npanel-pro/internal/service/auth"
)

type adapter struct {
	inner *data.AuthCompat
}

func New(inner *data.AuthCompat) authservice.AuthCompatProvider {
	if inner == nil {
		return nil
	}
	return &adapter{inner: inner}
}

func (a *adapter) GenerateCaptcha(ctx context.Context) (*authservice.CompatGenerateCaptchaResult, error) {
	result, err := a.inner.GenerateCaptcha(ctx)
	if err != nil {
		return nil, err
	}
	return &authservice.CompatGenerateCaptchaResult{
		ID:         result.ID,
		Image:      result.Image,
		Type:       result.Type,
		BlockImage: result.BlockImage,
	}, nil
}

func (a *adapter) VerifySliderCaptcha(ctx context.Context, id string, x, y int, trail string) (*authservice.CompatSliderVerifyResult, error) {
	result, err := a.inner.VerifySliderCaptcha(ctx, id, x, y, trail)
	if err != nil {
		return nil, err
	}
	return &authservice.CompatSliderVerifyResult{Token: result.Token}, nil
}

func (a *adapter) DeviceLogin(ctx context.Context, params *authservice.CompatDeviceLoginParams) (*authbiz.LoginResult, error) {
	if params == nil {
		return a.inner.DeviceLogin(ctx, nil)
	}
	return a.inner.DeviceLogin(ctx, &data.DeviceLoginParams{
		Identifier: params.Identifier,
		ShortCode:  params.ShortCode,
		Meta:       params.Meta,
	})
}

func (a *adapter) AdminLogin(ctx context.Context, params *authservice.CompatAdminLoginParams) (*authbiz.LoginResult, error) {
	if params == nil {
		return a.inner.AdminLogin(ctx, nil)
	}
	return a.inner.AdminLogin(ctx, &data.AdminLoginParams{
		Email:    params.Email,
		Password: params.Password,
		Meta:     params.Meta,
	})
}

func (a *adapter) AdminResetPassword(ctx context.Context, params *authservice.CompatAdminResetPasswordParams) (*authbiz.LoginResult, error) {
	if params == nil {
		return a.inner.AdminResetPassword(ctx, nil)
	}
	return a.inner.AdminResetPassword(ctx, &data.AdminResetPasswordParams{
		Email:    params.Email,
		Password: params.Password,
		Code:     params.Code,
		Meta:     params.Meta,
	})
}
