package captcha

import (
	"context"
	"strings"

	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/redis/go-redis/v9"
)

type VerifyInput struct {
	CaptchaID   string
	CaptchaCode string
	CfToken     string
	SliderToken string
	IP          string
}

func VerifyCaptcha(ctx context.Context, redisClient *redis.Client, captchaType string, turnstileSecret string, input VerifyInput) error {
	switch strings.ToLower(strings.TrimSpace(captchaType)) {
	case "", "none":
		return nil
	case string(CaptchaTypeLocal):
		svc := NewService(Config{Type: CaptchaTypeLocal, RedisClient: redisClient})
		ok, err := svc.Verify(ctx, input.CaptchaID, input.CaptchaCode, input.IP)
		if err != nil || !ok {
			return responsecode.NewKratosError(responsecode.ErrVerifyCodeError)
		}
	case string(CaptchaTypeTurnstile):
		svc := NewService(Config{Type: CaptchaTypeTurnstile, TurnstileSecret: turnstileSecret})
		ok, err := svc.Verify(ctx, input.CfToken, "", input.IP)
		if err != nil || !ok {
			return responsecode.NewKratosError(responsecode.ErrVerifyCodeError)
		}
	case string(CaptchaTypeSlider):
		svc := NewSliderService(redisClient)
		ok, err := svc.VerifySliderToken(ctx, input.SliderToken)
		if err != nil || !ok {
			return responsecode.NewKratosError(responsecode.ErrVerifyCodeError)
		}
	default:
		return nil
	}

	return nil
}
