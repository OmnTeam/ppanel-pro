package captcha

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/redis/go-redis/v9"
)

type CaptchaType string

const (
	CaptchaTypeLocal     CaptchaType = "local"
	CaptchaTypeTurnstile CaptchaType = "turnstile"
	CaptchaTypeSlider    CaptchaType = "slider"
)

type Service interface {
	Generate(ctx context.Context) (id string, image string, err error)
	Verify(ctx context.Context, token string, code string, ip string) (bool, error)
	GetType() CaptchaType
}

type SliderService interface {
	Service
	GenerateSlider(ctx context.Context) (id string, bgImage string, blockImage string, err error)
	VerifySlider(ctx context.Context, id string, x, y int, trail string) (token string, err error)
	VerifySliderToken(ctx context.Context, token string) (bool, error)
}

type Config struct {
	Type            CaptchaType
	RedisClient     *redis.Client
	TurnstileSecret string
}

func NewService(config Config) Service {
	switch config.Type {
	case CaptchaTypeTurnstile:
		return newTurnstileService(config.TurnstileSecret)
	case CaptchaTypeSlider:
		return newSliderService(config.RedisClient)
	case CaptchaTypeLocal:
		fallthrough
	default:
		return newLocalService(config.RedisClient)
	}
}

func NewSliderService(redisClient *redis.Client) SliderService {
	return newSliderService(redisClient)
}

type turnstileService struct {
	secret string
}

func newTurnstileService(secret string) Service {
	return &turnstileService{secret: secret}
}

func (s *turnstileService) Generate(ctx context.Context) (id string, image string, err error) {
	return "", "", nil
}

func (s *turnstileService) Verify(ctx context.Context, token string, code string, ip string) (bool, error) {
	if token == "" {
		return false, nil
	}
	return tool.VerifyTurnstileWithKey(ctx, s.secret, token, ip, tool.GenerateTurnstileIdempotencyKey())
}

func (s *turnstileService) GetType() CaptchaType {
	return CaptchaTypeTurnstile
}
