package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	"github.com/OmnTeam/ppanel-pro/ent/proxyservergroup"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxytrafficlog"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	servermodel "github.com/OmnTeam/ppanel-pro/internal/model/server"
	queueTypes "github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

var errCompatLegacyServerNotModified = errors.New("304 Not Modified")

type compatLegacyServerCommon struct {
	Protocol  string `form:"protocol"`
	ServerID  int64  `form:"server_id"`
	SecretKey string `form:"secret_key"`
}

type compatLegacyGetServerConfigRequest struct{ compatLegacyServerCommon }
type compatLegacyGetServerUserListRequest struct{ compatLegacyServerCommon }

type compatLegacyServerBasic struct {
	PullInterval int64 `json:"pull_interval"`
	PushInterval int64 `json:"push_interval"`
}

type compatLegacyGetServerConfigResponse struct {
	Basic    compatLegacyServerBasic `json:"basic"`
	Protocol string                  `json:"protocol"`
	Config   interface{}             `json:"config"`
}

type compatLegacyServerUser struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	SpeedLimit  int64  `json:"speed_limit"`
	DeviceLimit int64  `json:"device_limit"`
}

type compatLegacyGetServerUserListResponse struct {
	Users []compatLegacyServerUser `json:"users"`
}

type compatLegacyUserTraffic struct {
	SID      int64 `json:"uid"`
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
}

type compatLegacyPushUserTrafficRequest struct {
	compatLegacyServerCommon
	Traffic []compatLegacyUserTraffic `json:"traffic"`
}

type compatLegacyOnlineUser struct {
	SID int64  `json:"uid"`
	IP  string `json:"ip"`
}

type compatLegacyPushOnlineUsersRequest struct {
	compatLegacyServerCommon
	Users []compatLegacyOnlineUser `json:"users"`
}

type compatLegacyPushServerStatusRequest struct {
	compatLegacyServerCommon
	CPU       float64 `json:"cpu"`
	Mem       float64 `json:"mem"`
	Disk      float64 `json:"disk"`
	UpdatedAt int64   `json:"updated_at"`
}

type compatLegacyQueryServerConfigRequest struct {
	ServerID  int64
	SecretKey string   `form:"secret_key"`
	Protocols []string `form:"protocols,omitempty"`
}

type compatLegacyNodeDNS struct {
	Proto   string   `json:"proto"`
	Address string   `json:"address"`
	Domains []string `json:"domains"`
}

type compatLegacyNodeOutbound struct {
	Name     string   `json:"name"`
	Protocol string   `json:"protocol"`
	Address  string   `json:"address"`
	Port     int64    `json:"port"`
	Password string   `json:"password"`
	Rules    []string `json:"rules"`
}

type compatLegacyQueryServerConfigResponse struct {
	TrafficReportThreshold int64                      `json:"traffic_report_threshold"`
	IPStrategy             string                     `json:"ip_strategy"`
	DNS                    []compatLegacyNodeDNS      `json:"dns"`
	Block                  []string                   `json:"block"`
	Outbound               []compatLegacyNodeOutbound `json:"outbound"`
	Protocols              []*servermodel.Protocol    `json:"protocols"`
	Total                  int64                      `json:"total"`
}

type compatLegacyCodeError struct {
	code int
	msg  string
}

func (e *compatLegacyCodeError) Error() string { return e.msg }

func registerLegacyServerCompatRoutes(r *khttp.Router, dataLayer *data.Data) {
	if r == nil || dataLayer == nil || dataLayer.DB() == nil {
		return
	}

	r.GET("/v1/server/config", func(ctx khttp.Context) error {
		var req compatLegacyGetServerConfigRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if !compatLegacyServerSecretAllowed(ctx, dataLayer, req.SecretKey) {
			return ctx.String(http.StatusForbidden, "Forbidden")
		}
		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLegacyGetServerConfig(inner, dataLayer, request.(*compatLegacyGetServerConfigRequest))
		})
		if errors.Is(err, errCompatLegacyServerNotModified) {
			return ctx.String(http.StatusNotModified, "Not Modified")
		}
		if err != nil {
			return ctx.String(http.StatusNotFound, "Not Found")
		}
		return ctx.JSON(http.StatusOK, out)
	})

	r.GET("/v1/server/user", func(ctx khttp.Context) error {
		var req compatLegacyGetServerUserListRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if !compatLegacyServerSecretAllowed(ctx, dataLayer, req.SecretKey) {
			return ctx.String(http.StatusForbidden, "Forbidden")
		}
		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLegacyGetServerUserList(inner, dataLayer, request.(*compatLegacyGetServerUserListRequest))
		})
		if errors.Is(err, errCompatLegacyServerNotModified) {
			return ctx.String(http.StatusNotModified, "Not Modified")
		}
		if err != nil {
			return ctx.String(http.StatusNotFound, "Not Found")
		}
		return ctx.JSON(http.StatusOK, out)
	})

	r.POST("/v1/server/push", func(ctx khttp.Context) error {
		var req compatLegacyPushUserTrafficRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if !compatLegacyServerSecretAllowed(ctx, dataLayer, req.SecretKey) {
			return ctx.String(http.StatusForbidden, "Forbidden")
		}
		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatLegacyPushUserTraffic(inner, dataLayer, request.(*compatLegacyPushUserTrafficRequest))
		})
		if err != nil {
			return compatLegacyServerJSONError(ctx, err)
		}
		return compatLegacyServerJSON(ctx, nil)
	})

	r.POST("/v1/server/status", func(ctx khttp.Context) error {
		var req compatLegacyPushServerStatusRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if !compatLegacyServerSecretAllowed(ctx, dataLayer, req.SecretKey) {
			return ctx.String(http.StatusForbidden, "Forbidden")
		}
		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatLegacyPushServerStatus(inner, dataLayer, request.(*compatLegacyPushServerStatusRequest))
		})
		if err != nil {
			return compatLegacyServerJSONError(ctx, err)
		}
		return compatLegacyServerJSON(ctx, nil)
	})

	r.POST("/v1/server/online", func(ctx khttp.Context) error {
		var req compatLegacyPushOnlineUsersRequest
		_ = ctx.Bind(&req)
		_ = ctx.BindQuery(&req)
		if !compatLegacyServerSecretAllowed(ctx, dataLayer, req.SecretKey) {
			return ctx.String(http.StatusForbidden, "Forbidden")
		}
		_, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return nil, compatLegacyPushOnlineUsers(inner, dataLayer, request.(*compatLegacyPushOnlineUsersRequest))
		})
		if err != nil {
			return compatLegacyServerJSONError(ctx, err)
		}
		return compatLegacyServerJSON(ctx, nil)
	})

	r.GET("/v2/server/{server_id}", func(ctx khttp.Context) error {
		serverID, err := strconv.ParseInt(strings.TrimSpace(ctx.Vars().Get("server_id")), 10, 64)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid Params")
		}
		var req compatLegacyQueryServerConfigRequest
		req.ServerID = serverID
		if err := ctx.BindQuery(&req); err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid Params")
		}
		if !compatLegacyServerSecretAllowed(ctx, dataLayer, req.SecretKey) {
			return ctx.String(http.StatusUnauthorized, "Unauthorized")
		}
		out, err := compatMiddleware(ctx, &req, func(inner context.Context, request interface{}) (interface{}, error) {
			return compatLegacyQueryServerProtocolConfig(inner, dataLayer, request.(*compatLegacyQueryServerConfigRequest))
		})
		if err != nil {
			return compatLegacyServerJSONError(ctx, err)
		}
		return compatLegacyServerJSON(ctx, out)
	})
}

func compatLegacyServerJSON(ctx khttp.Context, data interface{}) error {
	return ctx.JSON(http.StatusOK, compatEnvelope{Code: 200, Msg: "success", Data: data})
}

func compatLegacyServerJSONError(ctx khttp.Context, err error) error {
	code := 500
	msg := "Internal Server Error"
	var typedErr *compatLegacyCodeError
	if errors.As(err, &typedErr) {
		code = typedErr.code
		msg = typedErr.msg
	}
	return ctx.JSON(http.StatusOK, compatEnvelope{Code: code, Msg: msg})
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

func compatLegacyServerSecretAllowed(ctx context.Context, dataLayer *data.Data, provided string) bool {
	if dataLayer == nil {
		return false
	}
	expected, err := compatSystemValue(ctx, dataLayer, "server", "NodeSecret", "node_secret")
	if err == nil && strings.TrimSpace(expected) != "" {
		return strings.TrimSpace(expected) == strings.TrimSpace(provided)
	}
	if appConf := dataLayer.AppConf(); appConf != nil && appConf.GetNode() != nil {
		expected = appConf.GetNode().GetNodeSecret()
	}
	return strings.TrimSpace(expected) != "" && strings.TrimSpace(expected) == strings.TrimSpace(provided)
}

func compatLegacyGetServerConfig(ctx context.Context, dataLayer *data.Data, req *compatLegacyGetServerConfigRequest) (*compatLegacyGetServerConfigResponse, error) {
	cacheKey := fmt.Sprintf("server:config:%d:%s", req.ServerID, req.Protocol)
	if dataLayer.Redis() != nil {
		if cache, err := dataLayer.Redis().Get(ctx, cacheKey).Result(); err == nil && cache != "" {
			etag := tool.GenerateETag([]byte(cache))
			if compatLegacyIfNoneMatch(ctx) == etag {
				return nil, errCompatLegacyServerNotModified
			}
			compatLegacySetReplyHeader(ctx, "ETag", etag)
			var resp compatLegacyGetServerConfigResponse
			if err := json.Unmarshal([]byte(cache), &resp); err != nil {
				return nil, err
			}
			return &resp, nil
		}
	}

	server, err := dataLayer.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}
	requestProtocol := req.Protocol
	if requestProtocol == "hysteria2" {
		requestProtocol = "hysteria"
	}
	protocols, err := servermodel.UnmarshalProtocols(server.Protocol)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	for _, protocol := range protocols {
		if protocol != nil && protocol.Type == requestProtocol {
			config = compatLegacyProtocolConfigMap(protocol)
			break
		}
	}

	resp := &compatLegacyGetServerConfigResponse{
		Basic: compatLegacyServerBasic{
			PullInterval: compatLegacySystemInt64(ctx, dataLayer, "server", 10, "NodePullInterval", "node_pull_interval"),
			PushInterval: compatLegacySystemInt64(ctx, dataLayer, "server", 60, "NodePushInterval", "node_push_interval"),
		},
		Protocol: req.Protocol,
		Config:   config,
	}
	encoded, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	etag := tool.GenerateETag(encoded)
	compatLegacySetReplyHeader(ctx, "ETag", etag)
	if dataLayer.Redis() != nil {
		_ = dataLayer.Redis().Set(ctx, cacheKey, encoded, -1).Err()
	}
	if compatLegacyIfNoneMatch(ctx) == etag {
		return nil, errCompatLegacyServerNotModified
	}
	return resp, nil
}

func compatLegacyGetServerUserList(ctx context.Context, dataLayer *data.Data, req *compatLegacyGetServerUserListRequest) (*compatLegacyGetServerUserListResponse, error) {
	cacheKey := fmt.Sprintf("server:user:%d", req.ServerID)
	if dataLayer.Redis() != nil {
		if cache, err := dataLayer.Redis().Get(ctx, cacheKey).Result(); err == nil && cache != "" {
			etag := tool.GenerateETag([]byte(cache))
			if compatLegacyIfNoneMatch(ctx) == etag {
				return nil, errCompatLegacyServerNotModified
			}
			compatLegacySetReplyHeader(ctx, "ETag", etag)
			var resp compatLegacyGetServerUserListResponse
			if err := json.Unmarshal([]byte(cache), &resp); err != nil {
				return nil, err
			}
			return &resp, nil
		}
	}

	if _, err := dataLayer.DB().ProxyServer.Get(ctx, req.ServerID); err != nil {
		return nil, err
	}
	nodes, err := dataLayer.DB().ProxyNode.Query().
		Where(proxynode.ServerIDEQ(req.ServerID), proxynode.ProtocolEQ(req.Protocol)).
		Order(ent.Asc(proxynode.FieldSort)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return compatLegacyDummyServerUserList(), nil
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

	subscribePlans, err := compatLegacyMatchedSubscribePlans(ctx, dataLayer, nodeGroupIDs, nodeIDs, tool.RemoveDuplicateElements(nodeTags...))
	if err != nil {
		return nil, err
	}
	if len(subscribePlans) == 0 {
		return compatLegacyDummyServerUserList(), nil
	}

	users := make([]compatLegacyServerUser, 0)
	now := time.Now()
	for _, subscribePlan := range subscribePlans {
		userSubs, err := compatLegacyUsersBySubscribeID(ctx, dataLayer, subscribePlan.ID)
		if err != nil {
			return nil, err
		}
		for _, userSub := range userSubs {
			if !compatLegacyShouldIncludeServerUser(ctx, dataLayer, userSub, nodeGroupIDs, now) {
				continue
			}
			users = append(users, compatLegacyServerUser{
				ID:          userSub.ID,
				UUID:        compatStringValue(userSub.UUID),
				SpeedLimit:  compatLegacyEffectiveSpeedLimit(ctx, dataLayer, subscribePlan, userSub, now),
				DeviceLimit: subscribePlan.DeviceLimit,
			})
		}
	}
	if len(nodeGroupIDs) > 0 {
		expiredUsers, expiredSpeedLimit, err := compatLegacyExpiredServerUsers(ctx, dataLayer, nodeGroupIDs)
		if err != nil {
			return nil, err
		}
		for i := range expiredUsers {
			if expiredSpeedLimit > 0 {
				expiredUsers[i].SpeedLimit = expiredSpeedLimit
			}
		}
		users = append(users, expiredUsers...)
	}
	if len(users) == 0 {
		return compatLegacyDummyServerUserList(), nil
	}

	resp := &compatLegacyGetServerUserListResponse{Users: users}
	encoded, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	etag := tool.GenerateETag(encoded)
	compatLegacySetReplyHeader(ctx, "ETag", etag)
	if dataLayer.Redis() != nil {
		_ = dataLayer.Redis().Set(ctx, cacheKey, encoded, -1).Err()
	}
	if compatLegacyIfNoneMatch(ctx) == etag {
		return nil, errCompatLegacyServerNotModified
	}
	return resp, nil
}

func compatLegacyPushUserTraffic(ctx context.Context, dataLayer *data.Data, req *compatLegacyPushUserTrafficRequest) error {
	server, err := dataLayer.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		return errors.New("server not found")
	}

	payload := queueTypes.TrafficStatistics{
		ServerID: server.ID,
		Protocol: req.Protocol,
		Logs:     make([]queueTypes.UserTraffic, 0, len(req.Traffic)),
	}
	for _, item := range req.Traffic {
		payload.Logs = append(payload.Logs, queueTypes.UserTraffic{
			SID:      item.SID,
			Upload:   item.Upload,
			Download: item.Download,
		})
	}

	if dataLayer.Queue() != nil {
		encoded, _ := json.Marshal(payload)
		task := asynq.NewTask(queueTypes.ForthwithTrafficStatistics, encoded, asynq.MaxRetry(3))
		_, _ = dataLayer.Queue().EnqueueContext(ctx, task)
	}

	now := time.Now()
	_ = dataLayer.DB().ProxyServer.UpdateOneID(server.ID).SetLastReportedAt(now).Exec(ctx)
	return nil
}

func compatLegacyPushServerStatus(ctx context.Context, dataLayer *data.Data, req *compatLegacyPushServerStatusRequest) error {
	server, err := dataLayer.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil || server.ID <= 0 {
		return errors.New("server not found")
	}

	if dataLayer.Redis() != nil {
		statusPayload := map[string]interface{}{
			"cpu":        req.CPU,
			"mem":        req.Mem,
			"disk":       req.Disk,
			"updated_at": req.UpdatedAt,
		}
		encoded, _ := json.Marshal(statusPayload)
		if err := dataLayer.Redis().Set(ctx, fmt.Sprintf("node:status:%d", req.ServerID), encoded, 5*time.Minute).Err(); err != nil {
			return errors.New("update node status failed")
		}
	}

	now := time.Now()
	_ = dataLayer.DB().ProxyServer.UpdateOneID(server.ID).SetLastReportedAt(now).Exec(ctx)
	return nil
}

func compatLegacyPushOnlineUsers(ctx context.Context, dataLayer *data.Data, req *compatLegacyPushOnlineUsersRequest) error {
	if req.ServerID <= 0 || len(req.Users) == 0 {
		return errors.New("invalid request parameters")
	}
	for _, user := range req.Users {
		if user.SID <= 0 || strings.TrimSpace(user.IP) == "" {
			return fmt.Errorf("invalid user data: uid=%d, ip=%s", user.SID, user.IP)
		}
	}
	if _, err := dataLayer.DB().ProxyServer.Get(ctx, req.ServerID); err != nil {
		return fmt.Errorf("server not found: %w", err)
	}
	if dataLayer.Redis() == nil {
		return nil
	}

	onlineUsers := make(map[int64][]string)
	for _, user := range req.Users {
		onlineUsers[user.SID] = append(onlineUsers[user.SID], user.IP)
	}

	encoded, err := json.Marshal(onlineUsers)
	if err != nil {
		return err
	}
	if err := dataLayer.Redis().Set(ctx, fmt.Sprintf("node:online:subscribe:%d:%s", req.ServerID, req.Protocol), encoded, 5*time.Minute).Err(); err != nil {
		return err
	}

	now := time.Now()
	expireAt := now.Add(5 * time.Minute).Unix()
	pipe := dataLayer.Redis().Pipeline()
	pipe.ZRemRangeByScore(ctx, "node:online:subscribe:global", "-inf", fmt.Sprintf("%d", now.Unix()))
	for subscribeID := range onlineUsers {
		pipe.ZAdd(ctx, "node:online:subscribe:global", redis.Z{Score: float64(expireAt), Member: subscribeID})
	}
	_, err = pipe.Exec(ctx)
	return err
}

func compatLegacyQueryServerProtocolConfig(ctx context.Context, dataLayer *data.Data, req *compatLegacyQueryServerConfigRequest) (*compatLegacyQueryServerConfigResponse, error) {
	server, err := dataLayer.DB().ProxyServer.Get(ctx, req.ServerID)
	if err != nil {
		return nil, err
	}

	protocols, err := servermodel.UnmarshalProtocols(server.Protocol)
	if err != nil {
		return nil, err
	}
	if len(req.Protocols) > 0 {
		requested := make(map[string]struct{}, len(req.Protocols))
		for _, item := range req.Protocols {
			requested[item] = struct{}{}
		}
		filtered := make([]*servermodel.Protocol, 0, len(protocols))
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

	dns := make([]compatLegacyNodeDNS, 0)
	if rawDNS := compatLegacySystemString(ctx, dataLayer, "server", "", "DNS", "dns"); rawDNS != "" {
		_ = json.Unmarshal([]byte(rawDNS), &dns)
	}
	block := make([]string, 0)
	if rawBlock := compatLegacySystemString(ctx, dataLayer, "server", "", "Block", "block"); rawBlock != "" {
		_ = json.Unmarshal([]byte(rawBlock), &block)
	}
	outbound := make([]compatLegacyNodeOutbound, 0)
	if rawOutbound := compatLegacySystemString(ctx, dataLayer, "server", "", "Outbound", "outbound"); rawOutbound != "" {
		_ = json.Unmarshal([]byte(rawOutbound), &outbound)
	}

	return &compatLegacyQueryServerConfigResponse{
		TrafficReportThreshold: compatLegacySystemInt64(ctx, dataLayer, "server", 0, "TrafficReportThreshold", "traffic_report_threshold"),
		IPStrategy:             compatLegacySystemString(ctx, dataLayer, "server", "", "IPStrategy", "ip_strategy"),
		DNS:                    dns,
		Block:                  block,
		Outbound:               outbound,
		Protocols:              protocols,
		Total:                  int64(len(protocols)),
	}, nil
}

func compatLegacyMatchedSubscribePlans(ctx context.Context, dataLayer *data.Data, nodeGroupIDs, nodeIDs []int64, nodeTags []string) ([]*ent.ProxySubscribe, error) {
	plans, err := dataLayer.DB().ProxySubscribe.Query().Order(ent.Asc(proxysubscribe.FieldSort)).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*ent.ProxySubscribe, 0, len(plans))
	for _, plan := range plans {
		if len(nodeGroupIDs) > 0 {
			if compatLegacySubscribeMatchesNodeGroups(plan, nodeGroupIDs) {
				result = append(result, plan)
			}
			continue
		}
		if compatLegacySubscribeMatchesNodesAndTags(plan, nodeIDs, nodeTags) {
			result = append(result, plan)
		}
	}
	return result, nil
}

func compatLegacySubscribeMatchesNodeGroups(plan *ent.ProxySubscribe, nodeGroupIDs []int64) bool {
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

func compatLegacySubscribeMatchesNodesAndTags(plan *ent.ProxySubscribe, nodeIDs []int64, nodeTags []string) bool {
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

func compatLegacyUsersBySubscribeID(ctx context.Context, dataLayer *data.Data, subscribeID int64) ([]*ent.ProxyUserSubscribe, error) {
	userSubs, err := dataLayer.DB().ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.SubscribeIDEQ(subscribeID), proxyusersubscribe.StatusIn(int8(0), int8(1))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := dataLayer.DB().ProxyUserSubscribe.Update().
		Where(proxyusersubscribe.SubscribeIDEQ(subscribeID), proxyusersubscribe.StatusEQ(int8(0))).
		SetStatus(int8(1)).
		Save(ctx); err != nil {
		return nil, err
	}
	return userSubs, nil
}

func compatLegacyShouldIncludeServerUser(ctx context.Context, dataLayer *data.Data, userSub *ent.ProxyUserSubscribe, nodeGroupIDs []int64, now time.Time) bool {
	if userSub == nil {
		return false
	}
	if compatIsLegacyUnlimitedTime(userSub.ExpireTime) {
		return true
	}
	if userSub.ExpireTime != nil && userSub.ExpireTime.After(now) {
		return true
	}
	return compatLegacyCanUseExpiredNodeGroup(ctx, dataLayer, userSub, nodeGroupIDs, now)
}

func compatLegacyExpiredServerUsers(ctx context.Context, dataLayer *data.Data, serverNodeGroupIDs []int64) ([]compatLegacyServerUser, int64, error) {
	expiredGroup, err := dataLayer.DB().ProxyServerGroup.Query().Where(proxyservergroup.IsExpiredGroupEQ(true)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, 0, nil
		}
		return nil, 0, err
	}
	if !tool.Contains(serverNodeGroupIDs, expiredGroup.ID) {
		return nil, 0, nil
	}

	userSubs, err := dataLayer.DB().ProxyUserSubscribe.Query().Where(proxyusersubscribe.StatusEQ(int8(3))).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	users := make([]compatLegacyServerUser, 0)
	seen := make(map[int64]struct{})
	now := time.Now()
	for _, userSub := range userSubs {
		if !compatLegacyExpiredUserEligible(userSub, expiredGroup, now) {
			continue
		}
		if _, ok := seen[userSub.ID]; ok {
			continue
		}
		seen[userSub.ID] = struct{}{}
		users = append(users, compatLegacyServerUser{ID: userSub.ID, UUID: compatStringValue(userSub.UUID)})
	}
	return users, int64(expiredGroup.SpeedLimit), nil
}

func compatLegacyExpiredUserEligible(userSub *ent.ProxyUserSubscribe, expiredGroup *ent.ProxyServerGroup, now time.Time) bool {
	if userSub == nil || expiredGroup == nil || userSub.ExpireTime == nil {
		return false
	}
	expiredDays := int(now.Sub(*userSub.ExpireTime).Hours() / 24)
	if expiredDays > expiredGroup.ExpiredDaysLimit {
		return false
	}
	if expiredGroup.MaxTrafficGBExpired != nil && *expiredGroup.MaxTrafficGBExpired > 0 {
		usedTrafficGB := float64(compatInt64Value(userSub.ExpiredDownload)+compatInt64Value(userSub.ExpiredUpload)) / (1024 * 1024 * 1024)
		if usedTrafficGB >= float64(*expiredGroup.MaxTrafficGBExpired) {
			return false
		}
	}
	return true
}

func compatLegacyCanUseExpiredNodeGroup(ctx context.Context, dataLayer *data.Data, userSub *ent.ProxyUserSubscribe, nodeGroupIDs []int64, now time.Time) bool {
	expiredGroup, err := dataLayer.DB().ProxyServerGroup.Query().Where(proxyservergroup.IsExpiredGroupEQ(true)).First(ctx)
	if err != nil {
		return false
	}
	if !tool.Contains(nodeGroupIDs, expiredGroup.ID) {
		return false
	}
	return compatLegacyExpiredUserEligible(userSub, expiredGroup, now)
}

func compatLegacyEffectiveSpeedLimit(ctx context.Context, dataLayer *data.Data, subscribePlan *ent.ProxySubscribe, userSub *ent.ProxyUserSubscribe, now time.Time) int64 {
	if subscribePlan == nil || userSub == nil {
		return 0
	}
	baseSpeedLimit := subscribePlan.SpeedLimit
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

		logs, err := dataLayer.DB().ProxyTrafficLog.Query().
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

func compatLegacyDummyServerUserList() *compatLegacyGetServerUserListResponse {
	return &compatLegacyGetServerUserListResponse{Users: []compatLegacyServerUser{{ID: 1, UUID: uuidx.NewUUID().String()}}}
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
		PaddingScheme:        config.PaddingScheme,
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
		result = compatLegacyAnyTLSNode{Port: uint16(config.Port), SecurityConfig: securityConfig}
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

func compatLegacyIfNoneMatch(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return strings.TrimSpace(tr.RequestHeader().Get("If-None-Match"))
	}
	return ""
}

func compatLegacySetReplyHeader(ctx context.Context, key, value string) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		tr.ReplyHeader().Set(key, value)
	}
}

func compatLegacySystemString(ctx context.Context, dataLayer *data.Data, category, fallback string, keys ...string) string {
	value, err := compatSystemValue(ctx, dataLayer, category, keys...)
	if err != nil || strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func compatLegacySystemInt64(ctx context.Context, dataLayer *data.Data, category string, fallback int64, keys ...string) int64 {
	value := compatLegacySystemString(ctx, dataLayer, category, "", keys...)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
