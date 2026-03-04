package server

import (
	"context"
	"encoding/json"

	v1 "github.com/OmnTeam/ppanel-pro/api/server/v1"
	serverBiz "github.com/OmnTeam/ppanel-pro/internal/biz/server"
	"github.com/go-kratos/kratos/v2/log"
)

// ServerService 节点服务器服务
type ServerService struct {
	v1.UnimplementedServerServer

	uc  *serverBiz.ServerNodeUsecase
	log *log.Helper
}

// NewServerService 创建节点服务器服务
func NewServerService(uc *serverBiz.ServerNodeUsecase, logger log.Logger) *ServerService {
	return &ServerService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// GetServerConfig 获取服务器配置
func (s *ServerService) GetServerConfig(ctx context.Context, req *v1.GetServerConfigRequest) (*v1.GetServerConfigReply, error) {
	config, err := s.uc.GetServerConfig(ctx, req.ServerId, req.Protocol, req.SecretKey)
	if err != nil {
		return nil, err
	}

	// 解析config JSON字符串为map
	var configMap map[string]string
	if err := json.Unmarshal([]byte(config.Config), &configMap); err != nil {
		s.log.Errorf("Failed to parse config: %v", err)
		configMap = make(map[string]string)
	}

	return &v1.GetServerConfigReply{
		Code:    0,
		Message: "success",
		Basic: &v1.ServerBasic{
			PushInterval: config.PushInterval,
			PullInterval: config.PullInterval,
		},
		Protocol: config.Protocol,
		Config:   configMap,
	}, nil
}

// GetServerUserList 获取服务器用户列表
func (s *ServerService) GetServerUserList(ctx context.Context, req *v1.GetServerUserListRequest) (*v1.GetServerUserListReply, error) {
	users, err := s.uc.GetServerUserList(ctx, req.ServerId, req.Protocol, req.SecretKey)
	if err != nil {
		return nil, err
	}

	userList := make([]*v1.ServerUser, 0, len(users))
	for _, user := range users {
		userList = append(userList, &v1.ServerUser{
			Id:          user.ID,
			Uuid:        user.UUID,
			SpeedLimit:  user.SpeedLimit,
			DeviceLimit: user.DeviceLimit,
		})
	}

	return &v1.GetServerUserListReply{
		Code:    0,
		Message: "success",
		Users:   userList,
	}, nil
}

// PushUserTraffic 推送用户流量
func (s *ServerService) PushUserTraffic(ctx context.Context, req *v1.PushUserTrafficRequest) (*v1.PushUserTrafficReply, error) {
	// 转换Traffic数据
	traffic := make([]*serverBiz.UserTraffic, 0, len(req.Traffic))
	for _, t := range req.Traffic {
		traffic = append(traffic, &serverBiz.UserTraffic{
			SID:      t.Sid,
			Upload:   t.Upload,
			Download: t.Download,
		})
	}

	bizReq := &serverBiz.PushUserTrafficRequest{
		ServerID:  req.ServerId,
		Protocol:  req.Protocol,
		SecretKey: req.SecretKey,
		Traffic:   traffic,
	}

	err := s.uc.PushUserTraffic(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	return &v1.PushUserTrafficReply{
		Code:    0,
		Message: "success",
	}, nil
}

// PushServerStatus 推送服务器状态
func (s *ServerService) PushServerStatus(ctx context.Context, req *v1.PushServerStatusRequest) (*v1.PushServerStatusReply, error) {
	bizReq := &serverBiz.PushServerStatusRequest{
		ServerID:  req.ServerId,
		Protocol:  req.Protocol,
		SecretKey: req.SecretKey,
		CPU:       req.Cpu,
		Mem:       req.Mem,
		Disk:      req.Disk,
		UpdatedAt: req.UpdatedAt,
	}

	err := s.uc.PushServerStatus(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	return &v1.PushServerStatusReply{
		Code:    0,
		Message: "success",
	}, nil
}

// PushOnlineUsers 推送在线用户
func (s *ServerService) PushOnlineUsers(ctx context.Context, req *v1.PushOnlineUsersRequest) (*v1.PushOnlineUsersReply, error) {
	// 转换Users数据
	users := make([]*serverBiz.OnlineUser, 0, len(req.Users))
	for _, u := range req.Users {
		users = append(users, &serverBiz.OnlineUser{
			SID: u.Sid,
			IP:  u.Ip,
		})
	}

	bizReq := &serverBiz.PushOnlineUsersRequest{
		ServerID:  req.ServerId,
		Protocol:  req.Protocol,
		SecretKey: req.SecretKey,
		Users:     users,
	}

	err := s.uc.PushOnlineUsers(ctx, bizReq)
	if err != nil {
		return nil, err
	}

	return &v1.PushOnlineUsersReply{
		Code:    0,
		Message: "success",
	}, nil
}

// QueryServerProtocolConfig 查询服务器协议配置
func (s *ServerService) QueryServerProtocolConfig(ctx context.Context, req *v1.QueryServerProtocolConfigRequest) (*v1.QueryServerProtocolConfigReply, error) {
	config, err := s.uc.QueryServerProtocolConfig(ctx, req.ServerId, req.SecretKey, req.Protocols)
	if err != nil {
		return nil, err
	}

	// 转换DNS配置
	dnsConfigs := make([]*v1.NodeDNS, 0, len(config.DNS))
	for _, dns := range config.DNS {
		dnsConfigs = append(dnsConfigs, &v1.NodeDNS{
			Server: dns.Server,
			Domain: dns.Domain,
			Port:   dns.Port,
		})
	}

	// 转换Outbound配置
	outboundConfigs := make([]*v1.NodeOutbound, 0, len(config.Outbound))
	for _, outbound := range config.Outbound {
		outboundConfigs = append(outboundConfigs, &v1.NodeOutbound{
			Tag:      outbound.Tag,
			Protocol: outbound.Protocol,
			Settings: outbound.Settings,
		})
	}

	// 转换Protocol配置
	protocolConfigs := make([]*v1.Protocol, 0, len(config.Protocols))
	for _, protocol := range config.Protocols {
		// 解析Config JSON字符串为map
		var configMap map[string]string
		if err := json.Unmarshal([]byte(protocol.Config), &configMap); err != nil {
			s.log.Errorf("Failed to parse protocol config: %v", err)
			configMap = make(map[string]string)
		}

		protocolConfigs = append(protocolConfigs, &v1.Protocol{
			Type:   protocol.Type,
			Config: configMap,
		})
	}

	return &v1.QueryServerProtocolConfigReply{
		Code:                   0,
		Message:                "success",
		TrafficReportThreshold: config.TrafficReportThreshold,
		IpStrategy:             config.IPStrategy,
		Dns:                    dnsConfigs,
		Block:                  config.Block,
		Outbound:               outboundConfigs,
		Protocols:              protocolConfigs,
		Total:                  config.Total,
	}, nil
}
