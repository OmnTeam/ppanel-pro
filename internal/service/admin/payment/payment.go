package payment

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/structpb"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/payment/v1"
	paymentbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/payment"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

// PaymentService 支付方式服务
type PaymentService struct {
	v1.UnimplementedPaymentServiceServer

	uc  *paymentbiz.PaymentUsecase
	log *log.Helper
}

// NewPaymentService 创建支付方式服务
func NewPaymentService(uc *paymentbiz.PaymentUsecase, logger log.Logger) *PaymentService {
	return &PaymentService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreatePaymentMethod 创建支付方式
func (s *PaymentService) CreatePaymentMethod(ctx context.Context, req *v1.CreatePaymentMethodRequest) (*v1.CreatePaymentMethodReply, error) {
	// 转换config为JSON字符串
	configJSON, err := tool.StructToJSON(req.Config)
	if err != nil {
		return nil, err
	}

	// 处理enable字段
	var enable *bool
	if req.Enable != nil {
		enable = req.Enable
	}

	method, err := s.uc.CreatePaymentMethod(
		ctx,
		req.Name,
		req.Platform,
		req.Description,
		req.Icon,
		req.Domain,
		configJSON,
		req.FeeMode,
		req.FeePercent,
		req.FeeAmount,
		enable,
	)
	if err != nil {
		return nil, err
	}

	// 转换config为Struct
	configStruct, err := s.parseConfig(method.Config)
	if err != nil {
		return nil, err
	}

	return &v1.CreatePaymentMethodReply{
		Code:    int32(responsecode.AdminCreatePaymentMethodSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreatePaymentMethodSuccess],
		Data: &v1.PaymentMethod{
			Id:          method.ID,
			TenantId:    method.TenantID,
			Name:        method.Name,
			Platform:    method.Platform,
			Description: method.Description,
			Icon:        method.Icon,
			Domain:      method.Domain,
			Config:      configStruct,
			FeeMode:     method.FeeMode,
			FeePercent:  method.FeePercent,
			FeeAmount:   method.FeeAmount,
			Enable:      method.Enable,
			NotifyUrl:   method.NotifyURL,
			Token:       method.Token,
		},
	}, nil
}

// UpdatePaymentMethod 更新支付方式
func (s *PaymentService) UpdatePaymentMethod(ctx context.Context, req *v1.UpdatePaymentMethodRequest) (*v1.UpdatePaymentMethodReply, error) {
	// 转换config为JSON字符串
	configJSON, err := tool.StructToJSON(req.Config)
	if err != nil {
		return nil, err
	}

	// 处理enable字段
	var enable *bool
	if req.Enable != nil {
		enable = req.Enable
	}

	method, err := s.uc.UpdatePaymentMethod(
		ctx,
		req.Id,
		req.Name,
		req.Platform,
		req.Description,
		req.Icon,
		req.Domain,
		configJSON,
		req.FeeMode,
		req.FeePercent,
		req.FeeAmount,
		enable,
	)
	if err != nil {
		return nil, err
	}

	// 转换config为Struct
	configStruct, err := s.parseConfig(method.Config)
	if err != nil {
		return nil, err
	}

	return &v1.UpdatePaymentMethodReply{
		Code:    int32(responsecode.AdminUpdatePaymentMethodSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdatePaymentMethodSuccess],
		Data: &v1.PaymentMethod{
			Id:          method.ID,
			TenantId:    method.TenantID,
			Name:        method.Name,
			Platform:    method.Platform,
			Description: method.Description,
			Icon:        method.Icon,
			Domain:      method.Domain,
			Config:      configStruct,
			FeeMode:     method.FeeMode,
			FeePercent:  method.FeePercent,
			FeeAmount:   method.FeeAmount,
			Enable:      method.Enable,
			NotifyUrl:   method.NotifyURL,
			Token:       method.Token,
		},
	}, nil
}

// DeletePaymentMethod 删除支付方式
func (s *PaymentService) DeletePaymentMethod(ctx context.Context, req *v1.DeletePaymentMethodRequest) (*v1.DeletePaymentMethodReply, error) {
	err := s.uc.DeletePaymentMethod(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.DeletePaymentMethodReply{
		Code:    int32(responsecode.AdminDeletePaymentMethodSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeletePaymentMethodSuccess],
	}, nil
}

// GetPaymentMethodList 获取支付方式列表
func (s *PaymentService) GetPaymentMethodList(ctx context.Context, req *v1.GetPaymentMethodListRequest) (*v1.GetPaymentMethodListReply, error) {
	// 处理enable字段
	var enable *bool
	if req.Enable != nil {
		enable = req.Enable
	}

	total, list, err := s.uc.GetPaymentMethodList(
		ctx,
		req.Page,
		req.Size,
		req.Platform,
		req.Search,
		enable,
	)
	if err != nil {
		return nil, err
	}

	methods := make([]*v1.PaymentMethod, 0, len(list))
	for _, method := range list {
		// 转换config为Struct
		configStruct, err := s.parseConfig(method.Config)
		if err != nil {
			return nil, err
		}

		methods = append(methods, &v1.PaymentMethod{
			Id:          method.ID,
			TenantId:    method.TenantID,
			Name:        method.Name,
			Platform:    method.Platform,
			Description: method.Description,
			Icon:        method.Icon,
			Domain:      method.Domain,
			Config:      configStruct,
			FeeMode:     method.FeeMode,
			FeePercent:  method.FeePercent,
			FeeAmount:   method.FeeAmount,
			Enable:      method.Enable,
			NotifyUrl:   method.NotifyURL,
			Token:       method.Token,
		})
	}

	return &v1.GetPaymentMethodListReply{
		Code:    int32(responsecode.AdminGetPaymentMethodListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetPaymentMethodListSuccess],
		Data: &v1.GetPaymentMethodListData{
			Total: total,
			List:  methods,
		},
	}, nil
}

// GetPaymentPlatform 获取支付平台列表
// 完全复刻原项目 server-master/internal/handler/admin/payment/getPaymentPlatformHandler.go
func (s *PaymentService) GetPaymentPlatform(ctx context.Context, req *v1.GetPaymentPlatformRequest) (*v1.GetPaymentPlatformReply, error) {
	platforms := s.uc.GetPaymentPlatform(ctx)

	platformList := make([]*v1.PaymentPlatform, 0, len(platforms))
	for _, platform := range platforms {
		platformList = append(platformList, &v1.PaymentPlatform{
			Platform:                 platform.Platform,
			PlatformUrl:              platform.PlatformUrl,
			PlatformFieldDescription: platform.PlatformFieldDescription,
		})
	}

	return &v1.GetPaymentPlatformReply{
		Code:    int32(responsecode.AdminGetPaymentPlatformSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetPaymentPlatformSuccess],
		Data: &v1.GetPaymentPlatformData{
			List: platformList,
		},
	}, nil
}

// parseConfig 解析JSON配置为Struct
func (s *PaymentService) parseConfig(configJSON string) (*structpb.Struct, error) {
	if configJSON == "" {
		return &structpb.Struct{}, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &data); err != nil {
		return &structpb.Struct{}, err
	}

	return structpb.NewStruct(data)
}
