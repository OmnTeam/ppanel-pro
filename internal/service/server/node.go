package server

import (
	"context"
	"encoding/json"

	v1 "github.com/OmnTeam/npanel-pro/api/server/v1"
	serverBiz "github.com/OmnTeam/npanel-pro/internal/biz/server"
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
	s.log.Infof(
		"[QueryServerProtocolConfig] request received server_id=%d secret_present=%t protocols=%v",
		req.ServerId,
		req.SecretKey != "",
		req.Protocols,
	)
	config, err := s.uc.QueryServerProtocolConfig(ctx, req.ServerId, req.SecretKey, req.Protocols)
	if err != nil {
		s.log.Errorf("[QueryServerProtocolConfig] usecase failed server_id=%d err=%v", req.ServerId, err)
		return nil, err
	}
	s.log.Infof(
		"[QueryServerProtocolConfig] usecase returned server_id=%d total=%d dns=%d outbound=%d protocols=%d",
		req.ServerId,
		config.Total,
		len(config.DNS),
		len(config.Outbound),
		len(config.Protocols),
	)

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
		if !protocol.Enable {
			continue
		}
		protocolConfigs = append(protocolConfigs, &v1.Protocol{
			Type:                          protocol.Type,
			Port:                          protocol.Port,
			Enable:                        protocol.Enable,
			Security:                      protocol.Security,
			Sni:                           protocol.SNI,
			AllowInsecure:                 protocol.AllowInsecure,
			Fingerprint:                   protocol.Fingerprint,
			RealityServerAddr:             protocol.RealityServerAddr,
			RealityServerPort:             protocol.RealityServerPort,
			RealityPrivateKey:             protocol.RealityPrivateKey,
			RealityPublicKey:              protocol.RealityPublicKey,
			RealityShortId:                protocol.RealityShortId,
			Transport:                     protocol.Transport,
			Host:                          protocol.Host,
			Path:                          protocol.Path,
			ServiceName:                   protocol.ServiceName,
			Cipher:                        protocol.Cipher,
			ServerKey:                     protocol.ServerKey,
			Flow:                          protocol.Flow,
			HopPorts:                      protocol.HopPorts,
			HopInterval:                   protocol.HopInterval,
			ObfsPassword:                  protocol.ObfsPassword,
			DisableSni:                    protocol.DisableSNI,
			ReduceRtt:                     protocol.ReduceRtt,
			UdpRelayMode:                  protocol.UDPRelayMode,
			CongestionController:          protocol.CongestionController,
			Multiplex:                     protocol.Multiplex,
			PaddingScheme:                 protocol.PaddingScheme,
			UpMbps:                        protocol.UpMbps,
			DownMbps:                      protocol.DownMbps,
			Obfs:                          protocol.Obfs,
			ObfsHost:                      protocol.ObfsHost,
			ObfsPath:                      protocol.ObfsPath,
			XhttpMode:                     protocol.XhttpMode,
			XhttpExtra:                    protocol.XhttpExtra,
			Encryption:                    protocol.Encryption,
			EncryptionMode:                protocol.EncryptionMode,
			EncryptionRtt:                 protocol.EncryptionRtt,
			EncryptionTicket:              protocol.EncryptionTicket,
			EncryptionServerPadding:       protocol.EncryptionServerPadding,
			EncryptionPrivateKey:          protocol.EncryptionPrivateKey,
			EncryptionClientPadding:       protocol.EncryptionClientPadding,
			EncryptionPassword:            protocol.EncryptionPassword,
			Ratio:                         protocol.Ratio,
			CertMode:                      protocol.CertMode,
			CertDnsProvider:               protocol.CertDNSProvider,
			CertDnsEnv:                    protocol.CertDNSEnv,
			SimnetPsk:                     protocol.SimnetPsk,
			SimnetKeyId:                   protocol.SimnetKeyID,
			SimnetTicketId:                protocol.SimnetTicketID,
			SimnetPath:                    protocol.SimnetPath,
			SimnetCarrier:                 protocol.SimnetCarrier,
			SimnetAfEnabled:               protocol.SimnetAfEnabled,
			SimnetAfPathMode:              protocol.SimnetAfPathMode,
			SimnetAfPathPrefix:            protocol.SimnetAfPathPrefix,
			SimnetAfPathSuffix:            protocol.SimnetAfPathSuffix,
			SimnetAfMagicMode:             protocol.SimnetAfMagicMode,
			SimnetAfResponseJitterMs:      protocol.SimnetAfResponseJitterMs,
			SimnetAfHandshakePolymorphism: protocol.SimnetAfHandshakePolymorphism,
			SimnetAfSettingsJitter:        protocol.SimnetAfSettingsJitter,
			SimnetAfFakeHeaderInjection:   protocol.SimnetAfFakeHeaderInjection,
			SimnetReverseEnabled:          protocol.SimnetReverseEnabled,
			SimnetReverseListenAddr:       protocol.SimnetReverseListenAddr,
			SimnetReverseListenPort:       protocol.SimnetReverseListenPort,
			SimnetReverseTargetHost:       protocol.SimnetReverseTargetHost,
			SimnetReverseTargetPort:       protocol.SimnetReverseTargetPort,
			SimnetFallbackEnabled:         protocol.SimnetFallbackEnabled,
			SimnetFallbackTargetScheme:    protocol.SimnetFallbackTargetScheme,
			SimnetFallbackTargetHost:      protocol.SimnetFallbackTargetHost,
			SimnetFallbackTargetPort:      protocol.SimnetFallbackTargetPort,
			SimnetFallbackHostHeader:      protocol.SimnetFallbackHostHeader,
			SimnetFallbackTlsSni:          protocol.SimnetFallbackTLSSNI,
			// OmniFlow
			OmniflowCarrier:                    protocol.OmniflowCarrier,
			OmniflowPath:                       protocol.OmniflowPath,
			OmniflowContentType:                protocol.OmniflowContentType,
			OmniflowProfilePath:                protocol.OmniflowProfilePath,
			OmniflowProfileJson:                protocol.OmniflowProfileJson,
			OmniflowServerHost:                 protocol.OmniflowServerHost,
			OmniflowServerPort:                 protocol.OmniflowServerPort,
			OmniflowCaCertPath:                 protocol.OmniflowCaCertPath,
			OmniflowTargetMeta:                 protocol.OmniflowTargetMeta,
			OmniflowSpkiPin:                    protocol.OmniflowSpkiPin,
			OmniflowH3FallbackEnabled:          protocol.OmniflowH3FallbackEnabled,
			OmniflowH3FallbackPolicy:           protocol.OmniflowH3FallbackPolicy,
			OmniflowH3FallbackTimeoutMs:        protocol.OmniflowH3FallbackTimeoutMs,
			OmniflowH3FallbackRetryBudget:      protocol.OmniflowH3FallbackRetryBudget,
			OmniflowH3FallbackSmokeEnabled:     protocol.OmniflowH3FallbackSmokeEnabled,
			OmniflowH3FallbackSmokeIntervalSec: protocol.OmniflowH3FallbackSmokeIntervalSec,
			OmniflowH3FallbackSmokeTimeoutMs:   protocol.OmniflowH3FallbackSmokeTimeoutMs,
			OmniflowMaxAgeSec:                  protocol.OmniflowMaxAgeSec,
			OmniflowIdleTimeoutSec:             protocol.OmniflowIdleTimeoutSec,
			OmniflowMaxConnections:             protocol.OmniflowMaxConnections,
			OmniflowAdaptiveTlsEnabled:         protocol.OmniflowAdaptiveTlsEnabled,
			OmniflowTlsFingerprint:             protocol.OmniflowTlsFingerprint,
			OmniflowSniMode:                    protocol.OmniflowSniMode,
			OmniflowPaddingMode:                protocol.OmniflowPaddingMode,
			OmniflowTrafficShapingEnabled:      protocol.OmniflowTrafficShapingEnabled,
			OmniflowFallbackCarrierEnabled:     protocol.OmniflowFallbackCarrierEnabled,
			OmniflowFallbackConnectTunnel:      protocol.OmniflowFallbackConnectTunnel,
			OmniflowFallbackWssEnabled:         protocol.OmniflowFallbackWssEnabled,
		})
	}

	reply := &v1.QueryServerProtocolConfigReply{
		Code:                   0,
		Message:                "success",
		TrafficReportThreshold: config.TrafficReportThreshold,
		IpStrategy:             config.IPStrategy,
		Dns:                    dnsConfigs,
		Block:                  config.Block,
		Outbound:               outboundConfigs,
		Protocols:              protocolConfigs,
		Total:                  int32(config.Total),
	}
	s.log.Infof(
		"[QueryServerProtocolConfig] reply ready server_id=%d total=%d dns=%d outbound=%d protocols=%d",
		req.ServerId,
		reply.Total,
		len(reply.Dns),
		len(reply.Outbound),
		len(reply.Protocols),
	)
	return reply, nil
}
