package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	"github.com/OmnTeam/ppanel-pro/ent/proxyservergroup"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	servermodel "github.com/OmnTeam/ppanel-pro/internal/model/server"
	queueTypes "github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type CompatLegacyProvider interface {
	DB() *ent.Client
	Redis() redis.UniversalClient
	Queue() *asynq.Client
	AppNodeConfig() *conf.Node
	LoadNodeConfig(ctx context.Context, module string) (*CompatLegacyNodeConfig, error)
}

type CompatLegacyNodeConfig struct {
	NodeSecret             string
	NodePullInterval       int64
	NodePushInterval       int64
	TrafficReportThreshold int64
	IPStrategy             string
	DNS                    string
	Block                  string
	Outbound               string
}

type CompatLegacyServerCommon struct {
	Protocol  string
	ServerID  int64
	SecretKey string
}

type CompatLegacyGetServerConfigRequest struct {
	CompatLegacyServerCommon
}

type CompatLegacyGetServerUserListRequest struct {
	CompatLegacyServerCommon
}

type CompatLegacyServerBasic struct {
	PushInterval int64 `json:"push_interval"`
	PullInterval int64 `json:"pull_interval"`
}

type CompatLegacyGetServerConfigResponse struct {
	Basic    CompatLegacyServerBasic `json:"basic"`
	Protocol string                  `json:"protocol"`
	Config   interface{}             `json:"config"`
}

type CompatLegacyServerUser struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  int64  `json:"speed_limit"`
	DeviceLimit int64  `json:"device_limit"`
}

type CompatLegacyGetServerUserListResponse struct {
	Users []CompatLegacyServerUser `json:"users"`
}

type CompatLegacyUserTraffic struct {
	SID      int64
	Upload   int64
	Download int64
}

type CompatLegacyPushUserTrafficRequest struct {
	CompatLegacyServerCommon
	Traffic []CompatLegacyUserTraffic
}

type CompatLegacyOnlineUser struct {
	SID int64
	IP  string
}

type CompatLegacyPushOnlineUsersRequest struct {
	CompatLegacyServerCommon
	Users []CompatLegacyOnlineUser
}

type CompatLegacyPushServerStatusRequest struct {
	CompatLegacyServerCommon
	CPU       float64
	Mem       float64
	Disk      float64
	UpdatedAt int64
}

type CompatLegacyQueryServerConfigRequest struct {
	ServerID  int64
	SecretKey string
	Protocols []string
}

type CompatLegacyNodeDNS struct {
	Proto   string   `json:"proto"`
	Address string   `json:"address"`
	Domains []string `json:"domains"`
}

type CompatLegacyNodeOutbound struct {
	Name     string   `json:"name"`
	Protocol string   `json:"protocol"`
	Address  string   `json:"address"`
	Port     int64    `json:"port"`
	Password string   `json:"password"`
	Rules    []string `json:"rules"`
}

type CompatLegacyQueryServerConfigResponse struct {
	TrafficReportThreshold int64                      `json:"traffic_report_threshold"`
	IPStrategy             string                     `json:"ip_strategy"`
	DNS                    []CompatLegacyNodeDNS      `json:"dns"`
	Block                  []string                   `json:"block"`
	Outbound               []CompatLegacyNodeOutbound `json:"outbound"`
	Protocols              []*servermodel.Protocol    `json:"protocols"`
	Total                  int64                      `json:"total"`
}

type compatLegacySecurityConfig struct {
	SNI                  string `json:"sni"`
	AllowInsecure        *bool  `json:"allow_insecure"`
	Fingerprint          string `json:"fingerprint"`
	RealityServerAddress string `json:"reality_server_addr"`
	RealityServerPort    int    `json:"reality_server_port"`
	RealityPrivateKey    string `json:"reality_private_key"`
	RealityPublicKey     string `json:"reality_public_key"`
	RealityShortID       string `json:"reality_short_id"`
	RealityMldsa65Seed   string `json:"reality_mldsa65seed"`
	PaddingScheme        string `json:"padding_scheme"`
}

type compatLegacyTransportConfig struct {
	Path                 string `json:"path"`
	Host                 string `json:"host"`
	ServiceName          string `json:"service_name"`
	DisableSNI           bool   `json:"disable_sni"`
	ReduceRtt            bool   `json:"reduce_rtt"`
	UDPRelayMode         string `json:"udp_relay_mode"`
	CongestionController string `json:"congestion_controller"`
}

type compatLegacyVlessNode struct {
	Port            uint16                       `json:"port"`
	Flow            string                       `json:"flow"`
	Network         string                       `json:"transport"`
	TransportConfig *compatLegacyTransportConfig `json:"transport_config"`
	Security        string                       `json:"security"`
	SecurityConfig  *compatLegacySecurityConfig  `json:"security_config"`
}

type compatLegacyVmessNode struct {
	Port            uint16                       `json:"port"`
	Network         string                       `json:"transport"`
	TransportConfig *compatLegacyTransportConfig `json:"transport_config"`
	Security        string                       `json:"security"`
	SecurityConfig  *compatLegacySecurityConfig  `json:"security_config"`
}

type compatLegacyShadowsocksNode struct {
	Port      uint16 `json:"port"`
	Cipher    string `json:"method"`
	ServerKey string `json:"server_key"`
}

type compatLegacyTrojanNode struct {
	Port            uint16                       `json:"port"`
	Network         string                       `json:"transport"`
	TransportConfig *compatLegacyTransportConfig `json:"transport_config"`
	Security        string                       `json:"security"`
	SecurityConfig  *compatLegacySecurityConfig  `json:"security_config"`
}

type compatLegacyAnyTLSNode struct {
	Port           uint16                      `json:"port"`
	SecurityConfig *compatLegacySecurityConfig `json:"security_config"`
}

type compatLegacyTuicNode struct {
	Port           uint16                      `json:"port"`
	SecurityConfig *compatLegacySecurityConfig `json:"security_config"`
}

type compatLegacyHysteriaNode struct {
	Port           uint16                      `json:"port"`
	HopPorts       string                      `json:"hop_ports"`
	HopInterval    int                         `json:"hop_interval"`
	ObfsPassword   string                      `json:"obfs_password"`
	SecurityConfig *compatLegacySecurityConfig `json:"security_config"`
}

type compatLegacyTrafficLimitRule struct {
	StatType     string `json:"stat_type"`
	StatValue    int64  `json:"stat_value"`
	TrafficUsage int64  `json:"traffic_usage"`
	SpeedLimit   int64  `json:"speed_limit"`
}

func compatLegacyServerUserListCacheKey(serverID int64) string {
	return fmt.Sprintf("server:user:%d", serverID)
}

func compatLegacyServerConfigCacheKey(serverID int64, protocol string) string {
	return fmt.Sprintf("server:config:%d:%s", serverID, protocol)
}

func (s *ServerService) CompatV1ServerSecretAllowed(ctx context.Context, provider CompatLegacyProvider, provided string) bool {
	if strings.TrimSpace(provided) == "" {
		return false
	}
	expected, ok := s.compatExpectedV1ServerSecret(ctx, provider)
	if !ok {
		return false
	}
	return strings.TrimSpace(provided) == strings.TrimSpace(expected)
}

func (s *ServerService) compatExpectedV1ServerSecret(ctx context.Context, provider CompatLegacyProvider) (string, bool) {
	if provider == nil {
		return "", false
	}
	if node := provider.AppNodeConfig(); node != nil {
		return node.NodeSecret, true
	}
	nodeConfig, err := provider.LoadNodeConfig(ctx, "service/server/compat/v1")
	if err != nil || nodeConfig == nil {
		return "", false
	}
	return nodeConfig.NodeSecret, true
}

func (s *ServerService) CompatV2ServerSecretAllowed(ctx context.Context, provider CompatLegacyProvider, provided string) bool {
	if provider == nil {
		s.log.Errorf("[QueryServerProtocolConfig] secret validation aborted: provider unavailable provided=%q", provided)
		return false
	}
	if node := provider.AppNodeConfig(); node != nil {
		expected := node.NodeSecret
		matched := strings.TrimSpace(expected) != "" && strings.TrimSpace(expected) == strings.TrimSpace(provided)
		s.log.Infof("[QueryServerProtocolConfig] secret compare source=runtime_node_config provided=%q expected=%q matched=%t", provided, expected, matched)
		return matched
	}
	nodeConfig, err := provider.LoadNodeConfig(ctx, "service/server/compat/v2")
	if err != nil || nodeConfig == nil {
		s.log.Errorf("[QueryServerProtocolConfig] secret load failed source=admin_node_config err=%v", err)
		return false
	}
	expected := nodeConfig.NodeSecret
	matched := strings.TrimSpace(expected) != "" && strings.TrimSpace(expected) == strings.TrimSpace(provided)
	s.log.Infof("[QueryServerProtocolConfig] secret compare source=admin_node_config_fallback provided=%q expected=%q matched=%t", provided, expected, matched)
	return matched
}

func (s *ServerService) CompatGetServerConfig(ctx context.Context, provider CompatLegacyProvider, req *CompatLegacyGetServerConfigRequest, ifNoneMatch string) (*CompatLegacyGetServerConfigResponse, string, bool, error) {
	if provider == nil || provider.DB() == nil || req == nil {
		return nil, "", false, errors.New("invalid provider")
	}
	if redisClient := provider.Redis(); redisClient != nil {
		cacheKey := compatLegacyServerConfigCacheKey(req.ServerID, req.Protocol)
		if cached, err := redisClient.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			etag := tool.GenerateETag([]byte(cached))
			if ifNoneMatch == etag {
				return nil, etag, true, nil
			}
			resp := &CompatLegacyGetServerConfigResponse{}
			if err := json.Unmarshal([]byte(cached), resp); err != nil {
				return nil, "", false, err
			}
			return resp, etag, false, nil
		}
	}

	server, err := provider.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		return nil, "", false, err
	}
	requestProtocol := req.Protocol
	if requestProtocol == "hysteria2" {
		requestProtocol = "hysteria"
	}
	protocols, err := servermodel.UnmarshalProtocols(server.Protocol)
	if err != nil {
		return nil, "", false, err
	}
	var config map[string]interface{}
	for _, protocol := range protocols {
		if protocol != nil && protocol.Type == requestProtocol {
			config = compatLegacyProtocolConfigMap(protocol)
			break
		}
	}
	pullInterval := int64(0)
	pushInterval := int64(0)
	if node := provider.AppNodeConfig(); node != nil {
		pullInterval = node.NodePullInterval
		pushInterval = node.NodePushInterval
	} else {
		nodeConfig, err := provider.LoadNodeConfig(ctx, "service/server/compat/v1")
		if err != nil || nodeConfig == nil {
			return nil, "", false, err
		}
		pullInterval = nodeConfig.NodePullInterval
		pushInterval = nodeConfig.NodePushInterval
	}

	resp := &CompatLegacyGetServerConfigResponse{
		Basic: CompatLegacyServerBasic{
			PullInterval: pullInterval,
			PushInterval: pushInterval,
		},
		Protocol: req.Protocol,
		Config:   config,
	}
	encoded, err := json.Marshal(resp)
	if err != nil {
		return nil, "", false, err
	}
	etag := tool.GenerateETag(encoded)
	if redisClient := provider.Redis(); redisClient != nil {
		_ = redisClient.Set(ctx, compatLegacyServerConfigCacheKey(req.ServerID, req.Protocol), encoded, -1).Err()
	}
	if ifNoneMatch == etag {
		return nil, etag, true, nil
	}
	return resp, etag, false, nil
}

func (s *ServerService) CompatGetServerUserList(ctx context.Context, provider CompatLegacyProvider, req *CompatLegacyGetServerUserListRequest, ifNoneMatch string) (*CompatLegacyGetServerUserListResponse, string, bool, error) {
	if provider == nil || provider.DB() == nil || req == nil {
		return nil, "", false, errors.New("invalid provider")
	}
	if redisClient := provider.Redis(); redisClient != nil {
		cacheKey := compatLegacyServerUserListCacheKey(req.ServerID)
		if cached, err := redisClient.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			etag := tool.GenerateETag([]byte(cached))
			if ifNoneMatch == etag {
				return nil, etag, true, nil
			}
			resp := &CompatLegacyGetServerUserListResponse{}
			if err := json.Unmarshal([]byte(cached), resp); err != nil {
				return nil, "", false, err
			}
			return resp, etag, false, nil
		}
	}

	if _, err := provider.DB().ProxyServer.Get(ctx, req.ServerID); err != nil {
		return nil, "", false, err
	}
	nodes, err := provider.DB().ProxyNode.Query().
		Where(proxynode.ServerIDEQ(req.ServerID), proxynode.ProtocolEQ(req.Protocol)).
		Order(ent.Asc(proxynode.FieldSort)).
		Limit(1000).
		All(ctx)
	if err != nil {
		return nil, "", false, err
	}
	if len(nodes) == 0 {
		return compatLegacyDummyServerUserList(), "", false, nil
	}

	nodeGroupMap := make(map[int64]struct{})
	nodeIDs := make([]int64, 0, len(nodes))
	nodeTags := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID)
		if node.Tags != "" {
			nodeTags = append(nodeTags, strings.Split(node.Tags, ",")...)
		}
		for _, groupID := range node.NodeGroupIds {
			if groupID > 0 {
				nodeGroupMap[groupID] = struct{}{}
			}
		}
	}
	nodeGroupIDs := make([]int64, 0, len(nodeGroupMap))
	for groupID := range nodeGroupMap {
		nodeGroupIDs = append(nodeGroupIDs, groupID)
	}

	subscribePlans, err := s.compatMatchedSubscribePlans(ctx, provider, nodeGroupIDs, nodeIDs, tool.RemoveDuplicateElements(nodeTags...))
	if err != nil {
		return nil, "", false, err
	}
	if len(subscribePlans) == 0 {
		return compatLegacyDummyServerUserList(), "", false, nil
	}

	users := make([]CompatLegacyServerUser, 0)
	now := time.Now()
	for _, subscribePlan := range subscribePlans {
		userSubs, err := s.compatUsersBySubscribeID(ctx, provider, subscribePlan.ID)
		if err != nil {
			return nil, "", false, err
		}
		for _, userSub := range userSubs {
			if !s.compatShouldIncludeServerUser(ctx, provider, userSub, nodeGroupIDs, now) {
				continue
			}
			users = append(users, CompatLegacyServerUser{
				ID:          userSub.ID,
				UUID:        compatStringValue(userSub.UUID),
				SpeedLimit:  s.compatEffectiveSpeedLimit(ctx, provider, subscribePlan, userSub, now),
				DeviceLimit: int64(subscribePlan.DeviceLimit),
			})
		}
	}
	if len(nodeGroupIDs) > 0 {
		expiredUsers, expiredSpeedLimit, err := s.compatExpiredServerUsers(ctx, provider, nodeGroupIDs)
		if err != nil {
			return nil, "", false, err
		}
		for i := range expiredUsers {
			if expiredSpeedLimit > 0 {
				expiredUsers[i].SpeedLimit = expiredSpeedLimit
			}
		}
		users = append(users, expiredUsers...)
	}
	if len(users) == 0 {
		return compatLegacyDummyServerUserList(), "", false, nil
	}

	resp := &CompatLegacyGetServerUserListResponse{Users: users}
	encoded, err := json.Marshal(resp)
	if err != nil {
		return nil, "", false, err
	}
	etag := tool.GenerateETag(encoded)
	if redisClient := provider.Redis(); redisClient != nil {
		_ = redisClient.Set(ctx, compatLegacyServerUserListCacheKey(req.ServerID), encoded, -1).Err()
	}
	if ifNoneMatch == etag {
		return nil, etag, true, nil
	}
	return resp, etag, false, nil
}

func (s *ServerService) CompatPushUserTraffic(ctx context.Context, provider CompatLegacyProvider, req *CompatLegacyPushUserTrafficRequest) error {
	if provider == nil || provider.DB() == nil || req == nil {
		return errors.New("invalid provider")
	}
	server, err := provider.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		return errors.New("server not found")
	}

	payload := queueTypes.TrafficStatistics{
		ServerID: server.ID,
		Protocol: req.Protocol,
		Logs:     make([]queueTypes.UserTraffic, 0, len(req.Traffic)),
	}
	for _, item := range req.Traffic {
		payload.Logs = append(payload.Logs, queueTypes.UserTraffic{SID: item.SID, Upload: item.Upload, Download: item.Download})
	}

	if provider.Queue() != nil {
		encoded, _ := json.Marshal(payload)
		task := asynq.NewTask(queueTypes.ForthwithTrafficStatistics, encoded, asynq.MaxRetry(3))
		_, _ = provider.Queue().EnqueueContext(ctx, task)
	}

	now := time.Now()
	_ = provider.DB().ProxyServer.UpdateOneID(server.ID).SetLastReportedAt(now).Exec(ctx)
	return nil
}

func (s *ServerService) CompatPushServerStatus(ctx context.Context, provider CompatLegacyProvider, req *CompatLegacyPushServerStatusRequest) error {
	if provider == nil || provider.DB() == nil || req == nil {
		return errors.New("invalid provider")
	}
	server, err := provider.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil || server.ID <= 0 {
		return errors.New("server not found")
	}

	if provider.Redis() != nil {
		statusPayload := map[string]interface{}{
			"cpu":        req.CPU,
			"mem":        req.Mem,
			"disk":       req.Disk,
			"updated_at": req.UpdatedAt,
		}
		encoded, _ := json.Marshal(statusPayload)
		if err := provider.Redis().Set(ctx, fmt.Sprintf("node:status:%d", req.ServerID), encoded, 5*time.Minute).Err(); err != nil {
			return errors.New("update node status failed")
		}
	}

	now := time.Now()
	_ = provider.DB().ProxyServer.UpdateOneID(server.ID).SetLastReportedAt(now).Exec(ctx)
	return nil
}

func (s *ServerService) CompatPushOnlineUsers(ctx context.Context, provider CompatLegacyProvider, req *CompatLegacyPushOnlineUsersRequest) error {
	if provider == nil || provider.DB() == nil || req == nil || req.ServerID <= 0 || len(req.Users) == 0 {
		return errors.New("invalid request parameters")
	}
	for i := range req.Users {
		normalizedIP, ok := compatNormalizeOnlineUserIP(req.Users[i].IP)
		if req.Users[i].SID <= 0 || !ok {
			return fmt.Errorf("invalid user data: uid=%d, ip=%s", req.Users[i].SID, req.Users[i].IP)
		}
		req.Users[i].IP = normalizedIP
	}
	if _, err := provider.DB().ProxyServer.Get(ctx, req.ServerID); err != nil {
		return fmt.Errorf("server not found: %w", err)
	}
	if provider.Redis() == nil {
		return nil
	}

	onlineUsers := make(map[int64][]string)
	for _, user := range req.Users {
		onlineUsers[user.SID] = compatAppendUniqueOnlineUserIP(onlineUsers[user.SID], user.IP)
	}

	encoded, err := json.Marshal(onlineUsers)
	if err != nil {
		return err
	}
	if err := provider.Redis().Set(ctx, fmt.Sprintf("node:online:subscribe:%d:%s", req.ServerID, req.Protocol), encoded, 5*time.Minute).Err(); err != nil {
		return err
	}

	now := time.Now()
	expireAt := now.Add(5 * time.Minute).Unix()
	pipe := provider.Redis().Pipeline()
	pipe.ZRemRangeByScore(ctx, "node:online:subscribe:global", "-inf", fmt.Sprintf("%d", now.Unix()))
	for subscribeID := range onlineUsers {
		pipe.ZAdd(ctx, "node:online:subscribe:global", redis.Z{Score: float64(expireAt), Member: subscribeID})
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (s *ServerService) CompatQueryServerProtocolConfig(ctx context.Context, provider CompatLegacyProvider, req *CompatLegacyQueryServerConfigRequest) (*CompatLegacyQueryServerConfigResponse, error) {
	if provider == nil || provider.DB() == nil || req == nil {
		return nil, errors.New("invalid provider")
	}
	server, err := provider.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}

	protocols, err := servermodel.UnmarshalProtocols(server.Protocol)
	if err != nil {
		return nil, err
	}
	var enabledProtocols []*servermodel.Protocol
	for _, protocol := range protocols {
		if protocol == nil || !protocol.Enable {
			continue
		}
		enabledProtocols = append(enabledProtocols, protocol)
	}
	protocols = enabledProtocols
	if len(req.Protocols) > 0 {
		requested := make(map[string]struct{}, len(req.Protocols))
		for _, item := range req.Protocols {
			if item = strings.TrimSpace(item); item != "" {
				requested[item] = struct{}{}
			}
		}
		if len(requested) > 0 {
			var filtered []*servermodel.Protocol
			for _, protocol := range protocols {
				if protocol == nil {
					continue
				}
				if _, ok := requested[protocol.Type]; ok {
					filtered = append(filtered, protocol)
				}
			}
			protocols = filtered
		}
	}

	trafficReportThreshold := int64(0)
	ipStrategy := ""
	var dns []CompatLegacyNodeDNS
	var block []string
	var outbound []CompatLegacyNodeOutbound

	if node := provider.AppNodeConfig(); node != nil {
		trafficReportThreshold = node.TrafficReportThreshold
		ipStrategy = node.IpStrategy
		for _, item := range node.Dns {
			if item == nil {
				continue
			}
			dns = append(dns, CompatLegacyNodeDNS{
				Proto:   item.Proto,
				Address: item.Address,
				Domains: item.Domains,
			})
		}
		block = node.Block
		for _, item := range node.Outbound {
			if item == nil {
				continue
			}
			outbound = append(outbound, CompatLegacyNodeOutbound{
				Name:     item.Name,
				Protocol: item.Protocol,
				Address:  item.Address,
				Port:     item.Port,
				Password: item.Password,
				Rules:    item.Rules,
			})
		}
	} else {
		nodeConfig, err := provider.LoadNodeConfig(ctx, "service/server/compat/v2")
		if err != nil || nodeConfig == nil {
			return nil, err
		}
		trafficReportThreshold = nodeConfig.TrafficReportThreshold
		ipStrategy = nodeConfig.IPStrategy
		if raw := strings.TrimSpace(nodeConfig.DNS); raw != "" {
			if err := json.Unmarshal([]byte(raw), &dns); err != nil {
				return nil, err
			}
		}
		if raw := strings.TrimSpace(nodeConfig.Block); raw != "" {
			if err := json.Unmarshal([]byte(raw), &block); err != nil {
				return nil, err
			}
		}
		if raw := strings.TrimSpace(nodeConfig.Outbound); raw != "" {
			if err := json.Unmarshal([]byte(raw), &outbound); err != nil {
				return nil, err
			}
		}
	}

	return normalizeCompatLegacyQueryServerConfigResponse(&CompatLegacyQueryServerConfigResponse{
		TrafficReportThreshold: trafficReportThreshold,
		IPStrategy:             ipStrategy,
		DNS:                    dns,
		Block:                  block,
		Outbound:               outbound,
		Protocols:              protocols,
		Total:                  int64(len(protocols)),
	}), nil
}

func (s *ServerService) compatMatchedSubscribePlans(ctx context.Context, provider CompatLegacyProvider, nodeGroupIDs, nodeIDs []int64, nodeTags []string) ([]*ent.ProxySubscribe, error) {
	plans, err := provider.DB().ProxySubscribe.Query().
		Order(ent.Asc(proxysubscribe.FieldSort)).
		Limit(9999).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*ent.ProxySubscribe, 0, len(plans))
	for _, plan := range plans {
		if len(nodeGroupIDs) > 0 {
			if compatSubscribeMatchesNodeGroups(plan, nodeGroupIDs) {
				result = append(result, plan)
			}
			continue
		}
		if compatSubscribeMatchesNodesAndTags(plan, nodeIDs, nodeTags) {
			result = append(result, plan)
		}
	}
	return result, nil
}

func compatSubscribeMatchesNodeGroups(plan *ent.ProxySubscribe, nodeGroupIDs []int64) bool {
	if plan == nil || len(nodeGroupIDs) == 0 {
		return false
	}
	if plan.NodeGroupID != nil && tool.Contains(nodeGroupIDs, *plan.NodeGroupID) {
		return true
	}
	for _, groupID := range plan.NodeGroupIds {
		if tool.Contains(nodeGroupIDs, groupID) {
			return true
		}
	}
	return false
}

func compatSubscribeMatchesNodesAndTags(plan *ent.ProxySubscribe, nodeIDs []int64, nodeTags []string) bool {
	if plan == nil {
		return false
	}
	if len(nodeIDs) == 0 && len(nodeTags) == 0 {
		return false
	}
	if len(nodeIDs) > 0 {
		planNodeIDs := tool.StringToInt64Slice(plan.Nodes)
		matched := false
		for _, nodeID := range nodeIDs {
			if tool.Contains(planNodeIDs, nodeID) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if len(nodeTags) > 0 && !compatNodeMatchesTags(plan.NodeTags, nodeTags) {
		return false
	}
	return true
}

func (s *ServerService) compatUsersBySubscribeID(ctx context.Context, provider CompatLegacyProvider, subscribeID int64) ([]*ent.ProxyUserSubscribe, error) {
	userSubs, err := provider.DB().ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.SubscribeIDEQ(subscribeID), proxyusersubscribe.StatusIn(int8(0), int8(1))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := provider.DB().ProxyUserSubscribe.Update().
		Where(proxyusersubscribe.SubscribeIDEQ(subscribeID), proxyusersubscribe.StatusEQ(int8(0))).
		SetStatus(int8(1)).
		Save(ctx); err != nil {
		return nil, err
	}
	return userSubs, nil
}

func (s *ServerService) compatShouldIncludeServerUser(ctx context.Context, provider CompatLegacyProvider, userSub *ent.ProxyUserSubscribe, nodeGroupIDs []int64, now time.Time) bool {
	if userSub == nil {
		return false
	}
	if userSub.ExpireTime == nil {
		return true
	}
	if compatIsLegacyUnlimitedTime(userSub.ExpireTime) {
		return true
	}
	if userSub.ExpireTime != nil && userSub.ExpireTime.After(now) {
		return true
	}
	return s.compatCanUseExpiredNodeGroup(ctx, provider, userSub, nodeGroupIDs, now)
}

func (s *ServerService) compatExpiredServerUsers(ctx context.Context, provider CompatLegacyProvider, serverNodeGroupIDs []int64) ([]CompatLegacyServerUser, int64, error) {
	expiredGroup, err := provider.DB().ProxyServerGroup.Query().Where(proxyservergroup.IsExpiredGroupEQ(true)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	if !tool.Contains(serverNodeGroupIDs, expiredGroup.ID) {
		return nil, 0, nil
	}

	userSubs, err := provider.DB().ProxyUserSubscribe.Query().Where(proxyusersubscribe.StatusEQ(int8(3))).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	users := make([]CompatLegacyServerUser, 0)
	seen := make(map[int64]struct{})
	now := time.Now()
	for _, userSub := range userSubs {
		if !compatExpiredUserEligible(userSub, expiredGroup, now) {
			continue
		}
		if _, ok := seen[userSub.ID]; ok {
			continue
		}
		seen[userSub.ID] = struct{}{}
		users = append(users, CompatLegacyServerUser{ID: userSub.ID, UUID: compatStringValue(userSub.UUID)})
	}
	return users, int64(expiredGroup.SpeedLimit), nil
}

func compatExpiredUserEligible(userSub *ent.ProxyUserSubscribe, expiredGroup *ent.ProxyServerGroup, now time.Time) bool {
	if userSub == nil || expiredGroup == nil || userSub.ExpireTime == nil {
		return false
	}
	expiredDays := int(now.Sub(*userSub.ExpireTime).Hours() / 24)
	if expiredDays > expiredGroup.ExpiredDaysLimit {
		return false
	}
	if expiredGroup.MaxTrafficGBExpired != nil && *expiredGroup.MaxTrafficGBExpired > 0 {
		usedTrafficGB := (compatInt64Value(userSub.ExpiredDownload) + compatInt64Value(userSub.ExpiredUpload)) / (1024 * 1024 * 1024)
		if usedTrafficGB >= compatInt64Value(expiredGroup.MaxTrafficGBExpired) {
			return false
		}
	}
	return true
}

func (s *ServerService) compatCanUseExpiredNodeGroup(ctx context.Context, provider CompatLegacyProvider, userSub *ent.ProxyUserSubscribe, nodeGroupIDs []int64, now time.Time) bool {
	expiredGroup, err := provider.DB().ProxyServerGroup.Query().Where(proxyservergroup.IsExpiredGroupEQ(true)).First(ctx)
	if err != nil {
		return false
	}
	if !tool.Contains(nodeGroupIDs, expiredGroup.ID) {
		return false
	}
	return compatExpiredUserEligible(userSub, expiredGroup, now)
}

func (s *ServerService) compatEffectiveSpeedLimit(ctx context.Context, provider CompatLegacyProvider, subscribePlan *ent.ProxySubscribe, userSub *ent.ProxyUserSubscribe, now time.Time) int64 {
	if subscribePlan == nil || userSub == nil {
		return 0
	}
	baseSpeedLimit := int64(subscribePlan.SpeedLimit)
	if subscribePlan.TrafficLimit == nil || strings.TrimSpace(*subscribePlan.TrafficLimit) == "" {
		return baseSpeedLimit
	}

	var rules []compatLegacyTrafficLimitRule
	if err := json.Unmarshal([]byte(*subscribePlan.TrafficLimit), &rules); err != nil {
		return baseSpeedLimit
	}

	for _, rule := range rules {
		var startTime time.Time
		var endTime time.Time
		switch rule.StatType {
		case "hour":
			if rule.StatValue <= 0 {
				continue
			}
			startTime = now.Add(-time.Duration(rule.StatValue) * time.Hour)
			endTime = now
		case "day":
			if rule.StatValue <= 0 {
				continue
			}
			startTime = now.AddDate(0, 0, -int(rule.StatValue))
			endTime = now
		default:
			continue
		}

		logs, err := provider.DB().ProxyTrafficLog.Query().
			Where(
				proxytrafficlog.UserIDEQ(userSub.UserID),
				proxytrafficlog.SubscribeIDEQ(userSub.ID),
				proxytrafficlog.TimestampGTE(startTime),
				proxytrafficlog.TimestampLT(endTime),
			).
			All(ctx)
		if err != nil {
			continue
		}

		var usedTraffic int64
		for _, item := range logs {
			usedTraffic += item.Upload + item.Download
		}
		usedGB := float64(usedTraffic) / (1024 * 1024 * 1024)
		if usedGB >= float64(rule.TrafficUsage) && rule.SpeedLimit > 0 {
			if baseSpeedLimit == 0 || rule.SpeedLimit < baseSpeedLimit {
				return rule.SpeedLimit
			}
		}
	}
	return baseSpeedLimit
}

func compatLegacyDummyServerUserList() *CompatLegacyGetServerUserListResponse {
	return &CompatLegacyGetServerUserListResponse{Users: []CompatLegacyServerUser{{ID: 1, UUID: uuidx.NewUUID().String()}}}
}

func compatLegacyProtocolConfigMap(config *servermodel.Protocol) map[string]interface{} {
	if config == nil {
		return nil
	}

	allowInsecure := config.AllowInsecure
	securityConfig := &compatLegacySecurityConfig{
		SNI:                  config.SNI,
		AllowInsecure:        &allowInsecure,
		Fingerprint:          config.Fingerprint,
		RealityServerAddress: config.RealityServerAddr,
		RealityServerPort:    int(config.RealityServerPort),
		RealityPrivateKey:    config.RealityPrivateKey,
		RealityPublicKey:     config.RealityPublicKey,
		RealityShortID:       config.RealityShortId,
	}
	transportConfig := &compatLegacyTransportConfig{
		Path:                 config.Path,
		Host:                 config.Host,
		ServiceName:          config.ServiceName,
		DisableSNI:           config.DisableSNI,
		ReduceRtt:            config.ReduceRtt,
		UDPRelayMode:         config.UDPRelayMode,
		CongestionController: config.CongestionController,
	}

	var result interface{}
	switch config.Type {
	case "shadowsocks":
		result = compatLegacyShadowsocksNode{Port: uint16(config.Port), Cipher: config.Cipher, ServerKey: base64.StdEncoding.EncodeToString([]byte(config.ServerKey))}
	case "vless":
		result = compatLegacyVlessNode{Port: uint16(config.Port), Flow: config.Flow, Network: config.Transport, TransportConfig: transportConfig, Security: config.Security, SecurityConfig: securityConfig}
	case "vmess":
		result = compatLegacyVmessNode{Port: uint16(config.Port), Network: config.Transport, TransportConfig: transportConfig, Security: config.Security, SecurityConfig: securityConfig}
	case "trojan":
		result = compatLegacyTrojanNode{Port: uint16(config.Port), Network: config.Transport, TransportConfig: transportConfig, Security: config.Security, SecurityConfig: securityConfig}
	case "anytls":
		anyTLSSecurityConfig := *securityConfig
		anyTLSSecurityConfig.PaddingScheme = config.PaddingScheme
		result = compatLegacyAnyTLSNode{Port: uint16(config.Port), SecurityConfig: &anyTLSSecurityConfig}
	case "tuic":
		result = compatLegacyTuicNode{Port: uint16(config.Port), SecurityConfig: securityConfig}
	case "hysteria", "hysteria2":
		result = compatLegacyHysteriaNode{Port: uint16(config.Port), HopPorts: config.HopPorts, HopInterval: int(config.HopInterval), ObfsPassword: config.ObfsPassword, SecurityConfig: securityConfig}
	}
	if result == nil {
		return nil
	}

	resp := make(map[string]interface{})
	payload, _ := json.Marshal(result)
	_ = json.Unmarshal(payload, &resp)
	return resp
}

func normalizeCompatLegacyQueryServerConfigResponse(resp *CompatLegacyQueryServerConfigResponse) *CompatLegacyQueryServerConfigResponse {
	if resp == nil {
		return nil
	}
	if len(resp.DNS) == 0 {
		resp.DNS = nil
	} else {
		for i := range resp.DNS {
			if len(resp.DNS[i].Domains) == 0 {
				resp.DNS[i].Domains = nil
			}
		}
	}
	if len(resp.Block) == 0 {
		resp.Block = nil
	}
	if len(resp.Outbound) == 0 {
		resp.Outbound = nil
	} else {
		for i := range resp.Outbound {
			if len(resp.Outbound[i].Rules) == 0 {
				resp.Outbound[i].Rules = nil
			}
		}
	}
	if len(resp.Protocols) == 0 {
		resp.Protocols = nil
	}
	return resp
}

func compatSanitizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func compatSanitizeDNSList(values []CompatLegacyNodeDNS) []CompatLegacyNodeDNS {
	if len(values) == 0 {
		return nil
	}
	result := make([]CompatLegacyNodeDNS, 0, len(values))
	for _, item := range values {
		item.Proto = strings.TrimSpace(item.Proto)
		item.Address = strings.TrimSpace(item.Address)
		item.Domains = compatSanitizeStringList(item.Domains)
		if item.Proto == "" && item.Address == "" && len(item.Domains) == 0 {
			continue
		}
		result = append(result, item)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func compatSanitizeOutboundList(values []CompatLegacyNodeOutbound) []CompatLegacyNodeOutbound {
	if len(values) == 0 {
		return nil
	}
	result := make([]CompatLegacyNodeOutbound, 0, len(values))
	for _, item := range values {
		item.Name = strings.TrimSpace(item.Name)
		item.Protocol = strings.TrimSpace(item.Protocol)
		item.Address = strings.TrimSpace(item.Address)
		item.Password = strings.TrimSpace(item.Password)
		item.Rules = compatSanitizeStringList(item.Rules)
		if item.Name == "" && item.Protocol == "" && item.Address == "" && item.Port == 0 && item.Password == "" && len(item.Rules) == 0 {
			continue
		}
		result = append(result, item)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func compatNormalizeOnlineUserIP(ip string) (string, bool) {
	normalizedIP := strings.TrimSpace(ip)
	if normalizedIP == "" || net.ParseIP(normalizedIP) == nil {
		return "", false
	}
	return normalizedIP, true
}

func compatAppendUniqueOnlineUserIP(ips []string, ip string) []string {
	for _, existingIP := range ips {
		if existingIP == ip {
			return ips
		}
	}
	return append(ips, ip)
}

func compatIsLegacyUnlimitedTime(value *time.Time) bool {
	return value != nil && value.Unix() == 0
}

func compatNodeMatchesTags(nodeTags string, tags []string) bool {
	for _, tag := range tags {
		for _, item := range strings.Split(nodeTags, ",") {
			if strings.TrimSpace(item) == tag {
				return true
			}
		}
	}
	return false
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
	return strings.TrimSpace(*value)
}
