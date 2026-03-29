package handler

import "context"

type groupRecalculator interface {
	RecalculateGroup(ctx context.Context, mode string, triggerType string) (int64, error)
}
