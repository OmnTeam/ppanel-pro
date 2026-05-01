package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/go-kratos/kratos/v2/log"
)

type compatPathTokenRequest struct {
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

func compatInt64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func compatStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func compatSystemValue(ctx context.Context, dataLayer *data.Data, category string, keys ...string) (string, error) {
	if dataLayer == nil || dataLayer.DB() == nil {
		return "", fmt.Errorf("data layer unavailable")
	}
	helper := log.NewHelper(log.With(log.DefaultLogger, "module", "server/compat/config"))

	entries, err := dataLayer.DB().ProxySystem.Query().
		Where(proxysystem.CategoryEQ(category)).
		Order(
			ent.Desc(proxysystem.FieldUpdatedAt),
			ent.Desc(proxysystem.FieldID),
		).
		All(ctx)
	if err != nil {
		return "", err
	}

	for _, lookupKey := range keys {
		for _, entry := range entries {
			if strings.TrimSpace(entry.Key) == strings.TrimSpace(lookupKey) {
				helper.Infof(
					"[compatSystemValue] category=%s lookup_keys=%v matched_key=%q matched_value=%q mode=exact",
					category,
					keys,
					entry.Key,
					entry.Value,
				)
				return entry.Value, nil
			}
		}
	}

	for _, lookupKey := range keys {
		normalizedLookup := compatNormalizeConfigKey(lookupKey)
		candidates := make([]string, 0)
		for _, entry := range entries {
			if compatNormalizeConfigKey(entry.Key) != normalizedLookup {
				continue
			}
			candidates = append(candidates, fmt.Sprintf("%s=%s", entry.Key, entry.Value))
		}
		if len(candidates) == 0 {
			continue
		}
		entry := candidates[0]
		parts := strings.SplitN(entry, "=", 2)
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		helper.Infof(
			"[compatSystemValue] category=%s lookup_keys=%v matched_key=%q matched_value=%q mode=normalized candidates=%v",
			category,
			keys,
			parts[0],
			value,
			candidates,
		)
		return value, nil
	}

	helper.Errorf("[compatSystemValue] category=%s lookup_keys=%v not found", category, keys)
	return "", fmt.Errorf("system config not found: %s", strings.Join(keys, ","))
}

func compatNormalizeConfigKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, "_", "")
	return key
}

func compatRequiredFieldError(typeName, fieldName string) error {
	return compatParamError(fmt.Sprintf(
		"Key: '%s.%s' Error:Field validation for '%s' failed on the 'required' tag",
		typeName,
		fieldName,
		fieldName,
	))
}
