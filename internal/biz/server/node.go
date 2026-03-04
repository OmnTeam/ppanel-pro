package server

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// ServerNodeRepo 节点服务器仓储接口
type ServerNodeRepo interface {
	GetServerConfig(ctx context.Context, serverID int64, protocol string) (*ServerConfig, error)
	GetServerUserList(ctx context.Context, serverID int64, protocol string) ([]*ServerUser, error)
	PushUserTraffic(ctx context.Context, req *PushUserTrafficRequest) error
	PushServerStatus(ctx context.Context, req *PushServerStatusRequest) error
	PushOnlineUsers(ctx context.Context, req *PushOnlineUsersRequest) error
	QueryServerProtocolConfig(ctx context.Context, serverID int64) (*ProtocolConfig, error)
}

// ServerConfig 服务器配置
type ServerConfig struct {
	PushInterval int64
	PullInterval int64
	Protocol     string
	Config       string
}

// ServerUser 服务器用户
type ServerUser struct {
	ID          int64
	UUID        string
	SpeedLimit  int64
	DeviceLimit int64
}

// UserTraffic 用户流量
type UserTraffic struct {
	SID      int64
	Upload   int64
	Download int64
}

// PushUserTrafficRequest 推送用户流量请求
type PushUserTrafficRequest struct {
	ServerID  int64
	Protocol  string
	SecretKey string
	Traffic   []*UserTraffic
}

// PushServerStatusRequest 推送服务器状态请求
type PushServerStatusRequest struct {
	ServerID  int64
	Protocol  string
	SecretKey string
	CPU       float64
	Mem       float64
	Disk      float64
	UpdatedAt int64
}

// OnlineUser 在线用户
type OnlineUser struct {
	SID int64
	IP  string
}

// PushOnlineUsersRequest 推送在线用户请求
type PushOnlineUsersRequest struct {
	ServerID  int64
	Protocol  string
	SecretKey string
	Users     []*OnlineUser
}

// Protocol 协议配置
type Protocol struct {
	Type   string
	Config string
}

// ProtocolConfig 协议配置响应
type ProtocolConfig struct {
	TrafficReportThreshold int64
	IPStrategy             string
	DNS                    []*DNSConfig
	Block                  []string
	Outbound               []*OutboundConfig
	Protocols              []*Protocol
	Total                  int64
}

// DNSConfig DNS配置
type DNSConfig struct {
	Server string
	Domain string
	Port   int64
}

// OutboundConfig 出站配置
type OutboundConfig struct {
	Tag      string
	Protocol string
	Settings map[string]string
}

// ServerNodeUsecase 节点服务器用例
type ServerNodeUsecase struct {
	repo   ServerNodeRepo
	logger *log.Helper
}

// NewServerNodeUsecase 创建节点服务器用例
func NewServerNodeUsecase(repo ServerNodeRepo, logger log.Logger) *ServerNodeUsecase {
	return &ServerNodeUsecase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// validateSecretKey 验证密钥
func (uc *ServerNodeUsecase) validateSecretKey(serverID int64, secretKey string) bool {
	// 简单的密钥验证逻辑
	// 实际生产环境应该使用更安全的方式
	expectedKey := fmt.Sprintf("server_%d", serverID)
	hash := md5.Sum([]byte(expectedKey))
	return hex.EncodeToString(hash[:]) == secretKey
}

// GetServerConfig 获取服务器配置
func (uc *ServerNodeUsecase) GetServerConfig(ctx context.Context, serverID int64, protocol, secretKey string) (*ServerConfig, error) {
	// 验证密钥
	if !uc.validateSecretKey(serverID, secretKey) {
		uc.logger.Errorf("Invalid secret key for server %d", serverID)
		return nil, responsecode.ErrUnauthorized()
	}

	config, err := uc.repo.GetServerConfig(ctx, serverID, protocol)
	if err != nil {
		uc.logger.Errorf("GetServerConfig failed: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	return config, nil
}

// GetServerUserList 获取服务器用户列表
func (uc *ServerNodeUsecase) GetServerUserList(ctx context.Context, serverID int64, protocol, secretKey string) ([]*ServerUser, error) {
	// 验证密钥
	if !uc.validateSecretKey(serverID, secretKey) {
		uc.logger.Errorf("Invalid secret key for server %d", serverID)
		return nil, responsecode.ErrUnauthorized()
	}

	users, err := uc.repo.GetServerUserList(ctx, serverID, protocol)
	if err != nil {
		uc.logger.Errorf("GetServerUserList failed: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	return users, nil
}

// PushUserTraffic 推送用户流量
func (uc *ServerNodeUsecase) PushUserTraffic(ctx context.Context, req *PushUserTrafficRequest) error {
	// 验证密钥
	if !uc.validateSecretKey(req.ServerID, req.SecretKey) {
		uc.logger.Errorf("Invalid secret key for server %d", req.ServerID)
		return responsecode.ErrUnauthorized()
	}

	err := uc.repo.PushUserTraffic(ctx, req)
	if err != nil {
		uc.logger.Errorf("PushUserTraffic failed: %v", err)
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	return nil
}

// PushServerStatus 推送服务器状态
func (uc *ServerNodeUsecase) PushServerStatus(ctx context.Context, req *PushServerStatusRequest) error {
	// 验证密钥
	if !uc.validateSecretKey(req.ServerID, req.SecretKey) {
		uc.logger.Errorf("Invalid secret key for server %d", req.ServerID)
		return responsecode.ErrUnauthorized()
	}

	err := uc.repo.PushServerStatus(ctx, req)
	if err != nil {
		uc.logger.Errorf("PushServerStatus failed: %v", err)
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	return nil
}

// PushOnlineUsers 推送在线用户
func (uc *ServerNodeUsecase) PushOnlineUsers(ctx context.Context, req *PushOnlineUsersRequest) error {
	// 验证密钥
	if !uc.validateSecretKey(req.ServerID, req.SecretKey) {
		uc.logger.Errorf("Invalid secret key for server %d", req.ServerID)
		return responsecode.ErrUnauthorized()
	}

	err := uc.repo.PushOnlineUsers(ctx, req)
	if err != nil {
		uc.logger.Errorf("PushOnlineUsers failed: %v", err)
		return responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	return nil
}

// QueryServerProtocolConfig 查询服务器协议配置
func (uc *ServerNodeUsecase) QueryServerProtocolConfig(ctx context.Context, serverID int64, secretKey string, protocols []string) (*ProtocolConfig, error) {
	// 验证密钥
	if !uc.validateSecretKey(serverID, secretKey) {
		uc.logger.Errorf("Invalid secret key for server %d", serverID)
		return nil, responsecode.ErrUnauthorized()
	}

	config, err := uc.repo.QueryServerProtocolConfig(ctx, serverID)
	if err != nil {
		uc.logger.Errorf("QueryServerProtocolConfig failed: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	return config, nil
}
