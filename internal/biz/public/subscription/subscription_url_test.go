package subscription

import (
	"context"
	"testing"
)

type subscribeURLRepoStub struct {
	path   string
	domain string
}

func (s *subscribeURLRepoStub) ValidateTokenAndGetSubscribe(ctx context.Context, token string) (*UserSubscribe, error) {
	return nil, nil
}
func (s *subscribeURLRepoStub) GetAvailableNodes(ctx context.Context, userSubscribe *UserSubscribe) ([]*NodeInfo, error) {
	return nil, nil
}
func (s *subscribeURLRepoStub) GetUserInfo(ctx context.Context, userID int64) (*UserInfo, error) {
	return nil, nil
}
func (s *subscribeURLRepoStub) GetSubscribeInfo(ctx context.Context, userSubscribe *UserSubscribe) string {
	return ""
}
func (s *subscribeURLRepoStub) UpdateSubscribeLog(ctx context.Context, userSubscribe *UserSubscribe, userAgent, clientIP string) error {
	return nil
}
func (s *subscribeURLRepoStub) GetSubscribeApplications(ctx context.Context) ([]*SubscribeApplication, error) {
	return nil, nil
}
func (s *subscribeURLRepoStub) GetSubscribeDomain(ctx context.Context) string {
	return s.domain
}
func (s *subscribeURLRepoStub) GetSubscribePath(ctx context.Context) string {
	return s.path
}
func (s *subscribeURLRepoStub) GetSiteName(ctx context.Context) string {
	return ""
}
func (s *subscribeURLRepoStub) GetSubscribeRuntimeConfig(ctx context.Context) (*SubscribeRuntimeConfig, error) {
	return nil, nil
}

func TestGetSubscribeV2URLUsesConfiguredPath(t *testing.T) {
	uc := NewSubscriptionUseCase(&subscribeURLRepoStub{
		path:   "/custom-subscribe",
		domain: "sub.example.com",
	})

	got := uc.getSubscribeV2URL(context.Background(), "demo-token", "/v1/subscribe/config?token=demo-token", "origin.example.com", false)
	want := "https://sub.example.com/custom-subscribe/demo-token"
	if got != want {
		t.Fatalf("getSubscribeV2URL() = %q, want %q", got, want)
	}
}

func TestGetSubscribeV2URLUsesGatewayPrefixWithConfiguredPath(t *testing.T) {
	uc := NewSubscriptionUseCase(&subscribeURLRepoStub{
		path: "/custom-subscribe",
	})

	got := uc.getSubscribeV2URL(context.Background(), "demo-token", "", "origin.example.com", true)
	want := "https://origin.example.com/sub/custom-subscribe/demo-token"
	if got != want {
		t.Fatalf("getSubscribeV2URL() = %q, want %q", got, want)
	}
}
