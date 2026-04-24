package server

import (
	"context"

	authbiz "github.com/OmnTeam/ppanel-pro/internal/biz/auth"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	authservice "github.com/OmnTeam/ppanel-pro/internal/service/auth"
)

type authCompatAdapter struct {
	inner *data.AuthCompat
}

func newAuthCompatAdapter(inner *data.AuthCompat) authservice.AuthCompatProvider {
	if inner == nil {
		return nil
	}
	return &authCompatAdapter{inner: inner}
}

func (a *authCompatAdapter) GenerateCaptcha(ctx context.Context) (*authservice.CompatGenerateCaptchaResult, error) {
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

func (a *authCompatAdapter) VerifySliderCaptcha(ctx context.Context, id string, x, y int, trail string) (*authservice.CompatSliderVerifyResult, error) {
	result, err := a.inner.VerifySliderCaptcha(ctx, id, x, y, trail)
	if err != nil {
		return nil, err
	}
	return &authservice.CompatSliderVerifyResult{Token: result.Token}, nil
}

func (a *authCompatAdapter) DeviceLogin(ctx context.Context, params *authservice.CompatDeviceLoginParams) (*authbiz.LoginResult, error) {
	if params == nil {
		return a.inner.DeviceLogin(ctx, nil)
	}
	return a.inner.DeviceLogin(ctx, &data.DeviceLoginParams{
		Identifier: params.Identifier,
		ShortCode:  params.ShortCode,
		Meta:       params.Meta,
	})
}
