package handler

import (
	"context"
	"encoding/json"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	authmodel "github.com/OmnTeam/ppanel-pro/internal/model/auth"
)

type queueEmailConfig struct {
	Platform                   string
	PlatformConfig             string
	VerifyEmailTemplate        string
	ExpirationEmailTemplate    string
	MaintenanceEmailTemplate   string
	TrafficExceedEmailTemplate string
}

type queueMobileConfig struct {
	Platform       string
	PlatformConfig string
}

func loadQueueEmailConfig(ctx context.Context, db *ent.Client) (*queueEmailConfig, error) {
	method, err := db.ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ("email")).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	var config authmodel.EmailAuthConfig
	config.Unmarshal(method.Config)

	platformConfig, err := json.Marshal(config.PlatformConfig)
	if err != nil {
		return nil, err
	}

	return &queueEmailConfig{
		Platform:                   config.Platform,
		PlatformConfig:             string(platformConfig),
		VerifyEmailTemplate:        config.VerifyEmailTemplate,
		ExpirationEmailTemplate:    config.ExpirationEmailTemplate,
		MaintenanceEmailTemplate:   config.MaintenanceEmailTemplate,
		TrafficExceedEmailTemplate: config.TrafficExceedEmailTemplate,
	}, nil
}

func loadQueueMobileConfig(ctx context.Context, db *ent.Client) (*queueMobileConfig, error) {
	method, err := db.ProxyAuthMethod.Query().
		Where(proxyauthmethod.MethodEQ("mobile")).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	var config authmodel.MobileAuthConfig
	config.Unmarshal(method.Config)

	platformConfig, err := json.Marshal(config.PlatformConfig)
	if err != nil {
		return nil, err
	}

	return &queueMobileConfig{
		Platform:       config.Platform,
		PlatformConfig: string(platformConfig),
	}, nil
}

func loadQueueSiteName(ctx context.Context, db *ent.Client) (string, error) {
	entries, err := db.ProxySystem.Query().
		Where(proxysystem.CategoryEQ("site")).
		All(ctx)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		switch entry.Key {
		case "SiteName", "site_name":
			return entry.Value, nil
		}
	}

	return "", nil
}
