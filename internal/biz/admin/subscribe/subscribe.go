package subscribe

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/subscribe/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/internal/model"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

const module = "biz/admin/subscribe"

func parseStringID(s string) (int, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return int(val), nil
}

func parseStringID64(s string) (int64, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return val, nil
}

// SubscribeUseCase subscribe use case
type SubscribeUseCase struct {
	repo SubscribeRepo
	log  *log.Helper
}

// NewSubscribeUseCase create subscribe use case
func NewSubscribeUseCase(repo SubscribeRepo, logger log.Logger) *SubscribeUseCase {
	return &SubscribeUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", module)),
	}
}

// SubscribeRepo subscribe repository interface
type SubscribeRepo interface {
	// Subscribe operations
	CreateSubscribe(ctx context.Context, sub *model.Subscribe) error
	GetSubscribeByID(ctx context.Context, id int) (*ent.ProxySubscribe, error)
	UpdateSubscribe(ctx context.Context, sub *model.Subscribe) error
	DeleteSubscribe(ctx context.Context, id int) error
	GetSubscribeList(ctx context.Context, req *model.SubscribeListParams) ([]*ent.ProxySubscribe, int64, error)
	CheckSubscribeInUse(ctx context.Context, subscribeID int) (bool, error)
	BatchDeleteSubscribe(ctx context.Context, ids []int) error
	GetSubscribeMinSort(ctx context.Context, ids []int) (int64, error)
	BatchUpdateSubscribeSort(ctx context.Context, subscribes []*ent.ProxySubscribe) error

	// Subscribe group operations
	CreateSubscribeGroup(ctx context.Context, group *model.SubscribeGroup) error
	GetSubscribeGroupByID(ctx context.Context, id int) (*ent.ProxySubscribeGroup, error)
	UpdateSubscribeGroup(ctx context.Context, group *model.SubscribeGroup) error
	DeleteSubscribeGroup(ctx context.Context, id int) error
	GetSubscribeGroupList(ctx context.Context) ([]*ent.ProxySubscribeGroup, int64, error)
	BatchDeleteSubscribeGroup(ctx context.Context, ids []int) error

	// User subscription query (for checking if subscribe is in use)
	GetActiveUserSubscriptionCount(ctx context.Context, subscribeID int) (int64, error)
	GetActiveUserSubscriptionCountByIDs(ctx context.Context, subscribeIDs []int64) (map[int64]int64, error)
}

// ==================== Subscribe Operations ====================

// CreateSubscribe create subscribe
func (uc *SubscribeUseCase) CreateSubscribe(ctx context.Context, req *v1.CreateSubscribeRequest) error {
	// Marshal discount to JSON
	discountJSON := ""
	if len(req.Discount) > 0 {
		data, err := json.Marshal(convertDiscountToModel(req.Discount))
		if err != nil {
			uc.log.WithContext(ctx).Errorw("msg", "Marshal discount failed", "error", err)
			return responsecode.NewKratosError(responsecode.ErrInternalError)
		}
		discountJSON = string(data)
	}

	// Convert nodes to comma-separated string
	nodesInt := make([]int, len(req.Nodes))
	for i, v := range req.Nodes {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		nodesInt[i] = parsed
	}
	nodesStr := int64SliceToString(nodesInt)
	nodeTagsStr := stringSliceToString(req.NodeTags)

	sub := &model.Subscribe{
		Name:           req.Name,
		Language:       req.Language,
		Description:    req.Description,
		UnitPrice:      req.UnitPrice,
		UnitTime:       req.UnitTime,
		Discount:       discountJSON,
		Replacement:    req.Replacement,
		Inventory:      req.Inventory,
		Traffic:        req.Traffic,
		SpeedLimit:     req.SpeedLimit,
		DeviceLimit:    req.DeviceLimit,
		Quota:          req.Quota,
		Nodes:          nodesStr,
		NodeTags:       nodeTagsStr,
		Show:           getBoolValue(req.Show, false),
		Sell:           getBoolValue(req.Sell, false),
		DeductionRatio: req.DeductionRatio,
		AllowDeduction: getBoolValue(req.AllowDeduction, true),
		ResetCycle:     req.ResetCycle,
		RenewalReset:   getBoolValue(req.RenewalReset, false),
	}

	if err := uc.repo.CreateSubscribe(ctx, sub); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "CreateSubscribe failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// UpdateSubscribe update subscribe
func (uc *SubscribeUseCase) UpdateSubscribe(ctx context.Context, req *v1.UpdateSubscribeRequest) error {
	// Check if subscribe exists
	id, err := parseStringID(req.Id)
	if err != nil {
		return err
	}

	_, err = uc.repo.GetSubscribeByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribe subscribe not found", "error", err, "id", req.Id)
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribe GetSubscribeByID error", "error", err, "id", req.Id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Marshal discount to JSON
	discountJSON := ""
	if len(req.Discount) > 0 {
		data, err := json.Marshal(convertDiscountToModel(req.Discount))
		if err != nil {
			uc.log.WithContext(ctx).Errorw("msg", "Marshal discount failed", "error", err)
			return responsecode.NewKratosError(responsecode.ErrInternalError)
		}
		discountJSON = string(data)
	}

	// Convert nodes to comma-separated string
	nodesInt := make([]int, len(req.Nodes))
	for i, v := range req.Nodes {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		nodesInt[i] = parsed
	}
	nodesStr := int64SliceToString(nodesInt)
	nodeTagsStr := stringSliceToString(req.NodeTags)

	sub := &model.Subscribe{
		ID:             int64(id),
		Name:           req.Name,
		Language:       req.Language,
		Description:    req.Description,
		UnitPrice:      req.UnitPrice,
		UnitTime:       req.UnitTime,
		Discount:       discountJSON,
		Replacement:    req.Replacement,
		Inventory:      req.Inventory,
		Traffic:        req.Traffic,
		SpeedLimit:     req.SpeedLimit,
		DeviceLimit:    req.DeviceLimit,
		Quota:          req.Quota,
		Nodes:          nodesStr,
		NodeTags:       nodeTagsStr,
		Show:           getBoolValue(req.Show, false),
		Sell:           getBoolValue(req.Sell, false),
		Sort:           req.Sort,
		DeductionRatio: req.DeductionRatio,
		AllowDeduction: getBoolValue(req.AllowDeduction, true),
		ResetCycle:     req.ResetCycle,
		RenewalReset:   getBoolValue(req.RenewalReset, false),
	}

	if err := uc.repo.UpdateSubscribe(ctx, sub); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribe failed", "error", err, "id", req.Id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// DeleteSubscribe delete subscribe
func (uc *SubscribeUseCase) DeleteSubscribe(ctx context.Context, id int) error {
	// Check if subscribe is in use by active user subscriptions
	inUse, err := uc.repo.CheckSubscribeInUse(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "DeleteSubscribe CheckSubscribeInUse error", "error", err, "id", id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	if inUse {
		uc.log.WithContext(ctx).Warnw("msg", "DeleteSubscribe subscribe is in use", "id", id)
		return responsecode.NewKratosError(responsecode.ErrSubscribeInUse)
	}

	if err := uc.repo.DeleteSubscribe(ctx, id); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "DeleteSubscribe failed", "error", err, "id", id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// BatchDeleteSubscribe batch delete subscribes
func (uc *SubscribeUseCase) BatchDeleteSubscribe(ctx context.Context, ids []int) error {
	// Check each subscribe if it's in use
	for _, id := range ids {
		inUse, err := uc.repo.CheckSubscribeInUse(ctx, id)
		if err != nil {
			uc.log.WithContext(ctx).Errorw("msg", "BatchDeleteSubscribe CheckSubscribeInUse error", "error", err, "id", id)
			return responsecode.NewKratosError(responsecode.ErrInternalError)
		}

		if inUse {
			uc.log.WithContext(ctx).Warnw("msg", "BatchDeleteSubscribe subscribe is in use", "id", id)
			return responsecode.NewKratosError(responsecode.ErrSubscribeInUse)
		}
	}

	// Delete all subscribes
	if err := uc.repo.BatchDeleteSubscribe(ctx, ids); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "BatchDeleteSubscribe failed", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// GetSubscribeDetails get subscribe details
func (uc *SubscribeUseCase) GetSubscribeDetails(ctx context.Context, id int) (*v1.SubscribeInfo, error) {
	sub, err := uc.repo.GetSubscribeByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeDetails subscribe not found", "error", err, "id", id)
			return nil, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeDetails failed", "error", err, "id", id)
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return convertSubscribeToProto(sub), nil
}

// GetSubscribeList get subscribe list
func (uc *SubscribeUseCase) GetSubscribeList(ctx context.Context, req *v1.GetSubscribeListRequest) (*v1.GetSubscribeListData, error) {
	params := &model.SubscribeListParams{
		Page:     int(req.Page),
		Size:     int(req.Size),
		Language: req.Language,
		Search:   req.Search,
	}

	list, total, err := uc.repo.GetSubscribeList(ctx, params)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeList failed", "error", err, "params", params)
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Get subscribe IDs for querying sold counts
	subscribeIDs := make([]int64, 0, len(list))
	for _, sub := range list {
		subscribeIDs = append(subscribeIDs, sub.ID)
	}

	// Get active user subscription counts (sold count)
	soldCounts := make(map[int64]int64)
	if len(subscribeIDs) > 0 {
		soldCounts, err = uc.repo.GetActiveUserSubscriptionCountByIDs(ctx, subscribeIDs)
		if err != nil {
			uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeList GetActiveUserSubscriptionCountByIDs error", "error", err)
			// Don't fail the request, just log the error
		}
	}

	// Convert to proto
	items := make([]*v1.SubscribeItem, 0, len(list))
	for _, sub := range list {
		item := convertSubscribeToProtoItem(sub)
		item.Sold = soldCounts[int64(sub.ID)]
		items = append(items, item)
	}

	return &v1.GetSubscribeListData{
		List:  items,
		Total: total,
	}, nil
}

// SubscribeSort subscribe sort
func (uc *SubscribeUseCase) SubscribeSort(ctx context.Context, req *v1.SubscribeSortRequest) error {
	if len(req.Sort) == 0 {
		return nil
	}

	// Extract IDs
	ids := make([]int, 0, len(req.Sort))
	sortMap := make(map[int64]int64)
	for i, item := range req.Sort {
		id, err := parseStringID(item.Id)
		if err != nil {
			return err
		}
		ids = append(ids, id)
		sortMap[int64(id)] = int64(i)
	}

	// Get minimum sort value
	minSort, err := uc.repo.GetSubscribeMinSort(ctx, ids)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "SubscribeSort GetSubscribeMinSort error", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Get subscribes
	idsInt64 := make([]int64, len(ids))
	for i, v := range ids {
		idsInt64[i] = int64(v)
	}
	params := &model.SubscribeListParams{
		Page: 1,
		Size: 9999,
		IDs:  idsInt64,
	}
	subscribes, _, err := uc.repo.GetSubscribeList(ctx, params)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "SubscribeSort GetSubscribeList error", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Update sort values
	for _, sub := range subscribes {
		if newSort, ok := sortMap[int64(sub.ID)]; ok {
			sub.Sort = int64(minSort + newSort)
		}
	}

	// Batch update
	if err := uc.repo.BatchUpdateSubscribeSort(ctx, subscribes); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "SubscribeSort BatchUpdateSubscribeSort error", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// ==================== Subscribe Group Operations ====================

// CreateSubscribeGroup create subscribe group
func (uc *SubscribeUseCase) CreateSubscribeGroup(ctx context.Context, req *v1.CreateSubscribeGroupRequest) error {
	group := &model.SubscribeGroup{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := uc.repo.CreateSubscribeGroup(ctx, group); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "CreateSubscribeGroup failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// UpdateSubscribeGroup update subscribe group
func (uc *SubscribeUseCase) UpdateSubscribeGroup(ctx context.Context, req *v1.UpdateSubscribeGroupRequest) error {
	id, err := parseStringID64(req.Id)
	if err != nil {
		return err
	}

	group := &model.SubscribeGroup{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := uc.repo.UpdateSubscribeGroup(ctx, group); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribeGroup failed", "error", err, "id", req.Id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// DeleteSubscribeGroup delete subscribe group
func (uc *SubscribeUseCase) DeleteSubscribeGroup(ctx context.Context, id int) error {
	if err := uc.repo.DeleteSubscribeGroup(ctx, id); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "DeleteSubscribeGroup failed", "error", err, "id", id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// BatchDeleteSubscribeGroup batch delete subscribe groups
func (uc *SubscribeUseCase) BatchDeleteSubscribeGroup(ctx context.Context, ids []int) error {
	if err := uc.repo.BatchDeleteSubscribeGroup(ctx, ids); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "BatchDeleteSubscribeGroup failed", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// GetSubscribeGroupList get subscribe group list
func (uc *SubscribeUseCase) GetSubscribeGroupList(ctx context.Context) (*v1.GetSubscribeGroupListData, error) {
	list, total, err := uc.repo.GetSubscribeGroupList(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeGroupList failed", "error", err)
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Convert to proto
	groups := make([]*v1.SubscribeGroupInfo, 0, len(list))
	for _, group := range list {
		desc := ""
		if group.Description != nil {
			desc = *group.Description
		}
		groups = append(groups, &v1.SubscribeGroupInfo{
			Id:          strconv.FormatInt(int64(group.ID), 10),
			Name:        group.Name,
			Description: desc,
			CreatedAt:   group.CreatedAt.Unix(),
			UpdatedAt:   group.UpdatedAt.Unix(),
		})
	}

	return &v1.GetSubscribeGroupListData{
		List:  groups,
		Total: total,
	}, nil
}

// ==================== Helper Functions ====================

// convertDiscountToModel convert proto discount to model discount
func convertDiscountToModel(discounts []*v1.SubscribeDiscount) []model.SubscribeDiscount {
	result := make([]model.SubscribeDiscount, 0, len(discounts))
	for _, d := range discounts {
		result = append(result, model.SubscribeDiscount{
			Quantity: d.Quantity,
			Discount: d.Discount,
		})
	}
	return result
}

// convertDiscountFromJSON convert JSON discount to proto discount
func convertDiscountFromJSON(discountJSON string) []*v1.SubscribeDiscount {
	if discountJSON == "" {
		return nil
	}

	var discounts []model.SubscribeDiscount
	if err := json.Unmarshal([]byte(discountJSON), &discounts); err != nil {
		return nil
	}

	result := make([]*v1.SubscribeDiscount, 0, len(discounts))
	for _, d := range discounts {
		result = append(result, &v1.SubscribeDiscount{
			Quantity: d.Quantity,
			Discount: d.Discount,
		})
	}
	return result
}

// int64SliceToString convert intslice to comma-separated string
func int64SliceToString(slice []int) string {
	if len(slice) == 0 {
		return ""
	}
	strs := make([]string, 0, len(slice))
	for _, v := range slice {
		strs = append(strs, fmt.Sprintf("%d", v))
	}
	return strings.Join(strs, ",")
}

// stringToInt64Slice convert comma-separated string to intslice
func stringToInt64Slice(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var val int
		fmt.Sscanf(p, "%d", &val)
		result = append(result, val)
	}
	return result
}

// stringSliceToString convert string slice to comma-separated string
func stringSliceToString(slice []string) string {
	return strings.Join(slice, ",")
}

// stringToStringSlice convert comma-separated string to string slice
func stringToStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// getBoolValue get bool value from optional bool pointer
func getBoolValue(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

// convertSubscribeToProto convert ent subscribe to proto subscribe info
func convertSubscribeToProto(sub *ent.ProxySubscribe) *v1.SubscribeInfo {
	desc := ""
	if sub.Description != nil {
		desc = *sub.Description
	}
	discount := ""
	if sub.Discount != nil {
		discount = *sub.Discount
	}
	deductionRatio := int64(0)
	if sub.DeductionRatio != nil {
		deductionRatio = int64(*sub.DeductionRatio)
	}
	allowDeduction := sub.AllowDeduction
	resetCycle := int64(0)
	if sub.ResetCycle != nil {
		resetCycle = int64(*sub.ResetCycle)
	}
	renewalReset := sub.RenewalReset

	return &v1.SubscribeInfo{
		Id:             strconv.FormatInt(int64(sub.ID), 10),
		Name:           sub.Name,
		Language:       sub.Language,
		Description:    desc,
		UnitPrice:      int64(sub.UnitPrice),
		UnitTime:       "", // sub.UnitTime is a string like "1m", "1h", needs parsing
		Discount:       convertDiscountFromJSON(discount),
		Replacement:    int64(sub.Replacement),
		Inventory:      int64(sub.Inventory),
		Traffic:        int64(sub.Traffic),
		SpeedLimit:     int64(sub.SpeedLimit),
		DeviceLimit:    int64(sub.DeviceLimit),
		Quota:          int64(sub.Quota),
		Nodes:          intSliceToStringSlice(stringToInt64Slice(sub.Nodes)),
		NodeTags:       stringToStringSlice(sub.NodeTags),
		Show:           sub.Show,
		Sell:           sub.Sell,
		Sort:           int64(sub.Sort),
		DeductionRatio: deductionRatio,
		AllowDeduction: allowDeduction,
		ResetCycle:     resetCycle,
		RenewalReset:   renewalReset,
		CreatedAt:      sub.CreatedAt.Unix(),
		UpdatedAt:      sub.UpdatedAt.Unix(),
	}
}

// convertSubscribeToProtoItem convert ent subscribe to proto subscribe item
func convertSubscribeToProtoItem(sub *ent.ProxySubscribe) *v1.SubscribeItem {
	desc := ""
	if sub.Description != nil {
		desc = *sub.Description
	}
	discount := ""
	if sub.Discount != nil {
		discount = *sub.Discount
	}
	deductionRatio := int64(0)
	if sub.DeductionRatio != nil {
		deductionRatio = int64(*sub.DeductionRatio)
	}
	allowDeduction := sub.AllowDeduction
	resetCycle := int64(0)
	if sub.ResetCycle != nil {
		resetCycle = int64(*sub.ResetCycle)
	}
	renewalReset := sub.RenewalReset

	return &v1.SubscribeItem{
		Id:             strconv.FormatInt(int64(sub.ID), 10),
		Name:           sub.Name,
		Language:       sub.Language,
		Description:    desc,
		UnitPrice:      int64(sub.UnitPrice),
		UnitTime:       "", // sub.UnitTime is a string like "1m", "1h", needs parsing
		Discount:       convertDiscountFromJSON(discount),
		Replacement:    int64(sub.Replacement),
		Inventory:      int64(sub.Inventory),
		Traffic:        int64(sub.Traffic),
		SpeedLimit:     int64(sub.SpeedLimit),
		DeviceLimit:    int64(sub.DeviceLimit),
		Quota:          int64(sub.Quota),
		Nodes:          convertIntSliceToInt64Slice(stringToInt64Slice(sub.Nodes)),
		NodeTags:       stringToStringSlice(sub.NodeTags),
		Show:           sub.Show,
		Sell:           sub.Sell,
		Sort:           int64(sub.Sort),
		DeductionRatio: deductionRatio,
		AllowDeduction: allowDeduction,
		ResetCycle:     resetCycle,
		RenewalReset:   renewalReset,
		CreatedAt:      sub.CreatedAt.Unix(),
		UpdatedAt:      sub.UpdatedAt.Unix(),
		Sold:           0, // Will be set by caller
	}
}

// convertIntSliceToInt64Slice converts []int to []int64
func convertIntSliceToInt64Slice(input []int) []int64 {
	if input == nil {
		return nil
	}
	result := make([]int64, len(input))
	for i, v := range input {
		result[i] = int64(v)
	}
	return result
}

func intSliceToStringSlice(input []int) []string {
	if input == nil {
		return nil
	}
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = strconv.FormatInt(int64(v), 10)
	}
	return result
}
