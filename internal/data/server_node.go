package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	serverBiz "github.com/OmnTeam/ppanel-pro/internal/biz/server"
	"github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// serverNodeRepo 节点服务器数据仓储
type serverNodeRepo struct {
	data  *Data
	log   *log.Helper
	queue *asynq.Client
}

// NewServerNodeRepo 创建节点服务器数据仓储
func NewServerNodeRepo(data *Data, logger log.Logger) serverBiz.ServerNodeRepo {
	return &serverNodeRepo{
		data:  data,
		log:   log.NewHelper(logger),
		queue: data.queue,
	}
}

// GetServerConfig 获取服务器配置
func (r *serverNodeRepo) GetServerConfig(ctx context.Context, serverID int64, protocol string) (*serverBiz.ServerConfig, error) {
	// 查找服务器
	server, err := r.data.db.ProxyServer.Get(ctx, serverID)
	if err != nil {
		r.log.Errorf("GetServerConfig failed: %v", err)
		return nil, err
	}

	// 解析协议配置
	var protocolConfig map[string]interface{}
	if server.Protocol != "" {
		var protocols []map[string]interface{}
		if err := json.Unmarshal([]byte(server.Protocol), &protocols); err != nil {
			r.log.Errorf("Failed to unmarshal protocols: %v", err)
			return nil, err
		}
		// 查找指定协议的配置
		for _, p := range protocols {
			if p["type"] == protocol {
				protocolConfig = p
				break
			}
		}
	}

	// 获取协议配置的JSON字符串
	configJSON := "{}"
	if len(protocolConfig) > 0 {
		if b, err := json.Marshal(protocolConfig); err == nil {
			configJSON = string(b)
		}
	}

	return &serverBiz.ServerConfig{
		PushInterval: 60, // 默认60秒
		PullInterval: 60, // 默认60秒
		Protocol:     protocol,
		Config:       configJSON,
	}, nil
}

// GetServerUserList 获取服务器用户列表
func (r *serverNodeRepo) GetServerUserList(ctx context.Context, serverID int64, protocol string) ([]*serverBiz.ServerUser, error) {
	// 查找启用的用户订阅
	subscribes, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.Status(int8(1)), // 启用状态
		).
		All(ctx)
	if err != nil {
		r.log.Errorf("Query subscribes failed: %v", err)
		return nil, err
	}

	// 构建用户列表
	users := make([]*serverBiz.ServerUser, 0, len(subscribes))
	for _, sub := range subscribes {
		// 处理UUID字段（可能为nil）
		var uuid string
		if sub.UUID != nil {
			uuid = *sub.UUID
		}
		// 处理Token字段（可能为nil）
		var token string
		if sub.Token != nil {
			token = *sub.Token
		}
		// 使用token作为UUID（如果UUID为空）
		if uuid == "" && token != "" {
			uuid = token
		}

		users = append(users, &serverBiz.ServerUser{
			ID:          sub.ID,
			UUID:        uuid,
			SpeedLimit:  0, // 从订阅套餐获取
			DeviceLimit: 0, // 从订阅套餐获取
		})
	}

	return users, nil
}

// PushUserTraffic 推送用户流量（创建流量统计任务）
func (r *serverNodeRepo) PushUserTraffic(ctx context.Context, req *serverBiz.PushUserTrafficRequest) error {
	// 验证服务器是否存在
	_, err := r.data.db.ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		r.log.Errorf("PushUserTraffic failed: server not found: %v", err)
		return fmt.Errorf("server not found")
	}

	// 更新服务器最后上报时间
	now := time.Now()
	err = r.data.db.ProxyServer.UpdateOneID(req.ServerID).
		SetLastReportedAt(now).
		Exec(ctx)
	if err != nil {
		r.log.Errorf("Update server last_reported_at failed: %v", err)
	}

	// 创建流量统计任务并推送到队列
	// 转换请求格式到队列任务格式
	userTrafficLogs := make([]types.UserTraffic, 0, len(req.Traffic))
	for _, traffic := range req.Traffic {
		userTrafficLogs = append(userTrafficLogs, types.UserTraffic{
			SID:      traffic.SID,
			Upload:   traffic.Upload,
			Download: traffic.Download,
		})
	}

	payload := types.TrafficStatistics{
		ServerID: req.ServerID,
		Protocol: req.Protocol,
		Logs:     userTrafficLogs,
	}

	// 序列化任务负载
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		r.log.Errorf("PushUserTraffic: Failed to marshal payload: %v", err)
		return fmt.Errorf("failed to marshal traffic payload: %w", err)
	}

	// 创建任务并立即执行
	task := asynq.NewTask(types.ForthwithTrafficStatistics, payloadBytes)

	// 入队任务（立即执行）
	_, err = r.queue.Enqueue(task)
	if err != nil {
		r.log.Errorf("PushUserTraffic: Failed to enqueue task: %v", err)
		return fmt.Errorf("failed to enqueue traffic task: %w", err)
	}

	r.log.Infof("PushUserTraffic: serverID=%d, protocol=%s, traffic count=%d, task enqueued",
		req.ServerID, req.Protocol, len(req.Traffic))

	return nil
}

// PushServerStatus 推送服务器状态
func (r *serverNodeRepo) PushServerStatus(ctx context.Context, req *serverBiz.PushServerStatusRequest) error {
	// 验证服务器是否存在
	_, err := r.data.db.ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		r.log.Errorf("PushServerStatus failed: server not found: %v", err)
		return fmt.Errorf("server not found")
	}

	// 更新服务器最后上报时间
	now := time.Now()
	err = r.data.db.ProxyServer.UpdateOneID(req.ServerID).
		SetLastReportedAt(now).
		Exec(ctx)
	if err != nil {
		r.log.Errorf("Update server last_reported_at failed: %v", err)
		return err
	}

	r.log.Infof("PushServerStatus: serverID=%d, cpu=%.2f, mem=%.2f, disk=%.2f",
		req.ServerID, req.CPU, req.Mem, req.Disk)

	return nil
}

// PushOnlineUsers 推送在线用户
func (r *serverNodeRepo) PushOnlineUsers(ctx context.Context, req *serverBiz.PushOnlineUsersRequest) error {
	// 验证服务器是否存在
	_, err := r.data.db.ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		r.log.Errorf("PushOnlineUsers failed: server not found: %v", err)
		return fmt.Errorf("server not found")
	}

	// 构建在线用户映射 map[subscribeID][]IP
	onlineUsers := make(map[int64][]string)
	for _, user := range req.Users {
		if ips, ok := onlineUsers[user.SID]; ok {
			// 如果用户已存在，追加IP
			onlineUsers[user.SID] = append(ips, user.IP)
		} else {
			// 新用户
			onlineUsers[user.SID] = []string{user.IP}
		}
	}

	// 存储到Redis缓存
	// 格式：node:online:subscribe:{serverID}:{protocol}
	key := fmt.Sprintf("node:online:subscribe:%d:%s", req.ServerID, req.Protocol)

	// 序列化在线用户数据
	data, err := json.Marshal(onlineUsers)
	if err != nil {
		r.log.Errorf("Marshal online users failed: %v", err)
		return err
	}

	// 存储到Redis（设置过期时间为5分钟）
	err = r.data.rdb.Set(ctx, key, data, 5*time.Minute).Err()
	if err != nil {
		r.log.Errorf("Redis Set failed: %v", err)
		return err
	}

	r.log.Infof("PushOnlineUsers: serverID=%d, protocol=%s, online users=%d",
		req.ServerID, req.Protocol, len(onlineUsers))

	return nil
}

// QueryServerProtocolConfig 查询服务器协议配置
func (r *serverNodeRepo) QueryServerProtocolConfig(ctx context.Context, serverID int64) (*serverBiz.ProtocolConfig, error) {
	// 查找服务器
	server, err := r.data.db.ProxyServer.Get(ctx, serverID)
	if err != nil {
		r.log.Errorf("QueryServerProtocolConfig failed: %v", err)
		return nil, err
	}

	// 解析协议配置
	var protocols []map[string]interface{}
	if server.Protocol != "" {
		if err := json.Unmarshal([]byte(server.Protocol), &protocols); err != nil {
			r.log.Errorf("Failed to unmarshal protocols: %v", err)
			return nil, err
		}
	}

	// 构建协议配置响应
	protocolConfigs := make([]*serverBiz.Protocol, 0, len(protocols))
	for _, p := range protocols {
		configJSON, _ := json.Marshal(p)
		protocolConfigs = append(protocolConfigs, &serverBiz.Protocol{
			Type:   p["type"].(string),
			Config: string(configJSON),
		})
	}

	return &serverBiz.ProtocolConfig{
		TrafficReportThreshold: 0,   // 从系统配置获取
		IPStrategy:             "",  // 从系统配置获取
		DNS:                    nil, // DNS配置
		Block:                  nil, // 阻止列表
		Outbound:               nil, // 出站配置
		Protocols:              protocolConfigs,
		Total:                  int64(len(protocolConfigs)),
	}, nil
}
