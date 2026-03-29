package captcha

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const localCaptchaTTL = 5 * time.Minute

type localService struct {
	redis *redis.Client
}

func newLocalService(redisClient *redis.Client) Service {
	return &localService{redis: redisClient}
}

func (s *localService) Generate(ctx context.Context) (id string, image string, err error) {
	if s.redis == nil {
		return "", "", fmt.Errorf("redis client is nil")
	}

	id = uuid.NewString()
	answer, err := randomCaptchaText(5)
	if err != nil {
		return "", "", err
	}

	key := localCaptchaCacheKey(id)
	if err := s.redis.Set(ctx, key, answer, localCaptchaTTL).Err(); err != nil {
		return "", "", err
	}

	svg := buildCaptchaSVG(answer)
	image = "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
	return id, image, nil
}

func (s *localService) Verify(ctx context.Context, id string, code string, ip string) (bool, error) {
	if s.redis == nil || id == "" || code == "" {
		return false, nil
	}

	key := localCaptchaCacheKey(id)
	answer, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	_ = s.redis.Del(ctx, key).Err()
	return strings.EqualFold(answer, code), nil
}

func (s *localService) GetType() CaptchaType {
	return CaptchaTypeLocal
}

func localCaptchaCacheKey(id string) string {
	return fmt.Sprintf("captcha:%s", id)
}

func randomCaptchaText(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789abcdefghijkmnpqrstuvwxyz"

	buf := make([]byte, length)
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	for i := range buf {
		buf[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(buf), nil
}

func buildCaptchaSVG(answer string) string {
	noise := []string{
		`<line x1="18" y1="18" x2="210" y2="54" stroke="#b0c4de" stroke-width="2"/>`,
		`<line x1="36" y1="62" x2="224" y2="22" stroke="#ffd59e" stroke-width="2"/>`,
		`<circle cx="32" cy="26" r="3" fill="#dbeafe"/>`,
		`<circle cx="214" cy="52" r="4" fill="#fce7f3"/>`,
	}

	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="240" height="80" viewBox="0 0 240 80">`+
			`<rect width="240" height="80" rx="10" fill="#f8fafc"/>`+
			`<rect x="4" y="4" width="232" height="72" rx="8" fill="none" stroke="#cbd5e1"/>`+
			`%s`+
			`<text x="24" y="54" font-size="34" font-family="monospace" letter-spacing="8" fill="#0f172a">%s</text>`+
			`</svg>`,
		strings.Join(noise, ""),
		answer,
	)
}
