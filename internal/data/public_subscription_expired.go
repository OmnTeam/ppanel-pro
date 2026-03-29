package data

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxynode"
	"github.com/OmnTeam/ppanel-pro/ent/proxyserver"
	"github.com/OmnTeam/ppanel-pro/ent/proxyservergroup"
	subscriptionbiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscription"
	servermodel "github.com/OmnTeam/ppanel-pro/internal/model/server"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

// createExpiredNodesFromDB 按照原项目逻辑从数据库获取过期节点组
func (r *publicSubscriptionRepo) createExpiredNodesFromDB(ctx context.Context, userSubscribe *subscriptionbiz.UserSubscribe) []*subscriptionbiz.NodeInfo {
	// 1. 查询过期节点组
	expiredGroup, err := r.data.db.ProxyServerGroup.Query().
		Where(proxyservergroup.IsExpiredGroupEQ(true)).
		First(ctx)

	if err != nil {
		r.log.Debugf("No expired node group configured: %v", err)
		return r.createExpiredNodesDefault()
	}

	// 2. 检查用户是否在过期天数限制内
	if userSubscribe.ExpireTime == 0 {
		r.log.Debug("User subscription has no expire time")
		return r.createExpiredNodesDefault()
	}

	expireTime := time.UnixMilli(userSubscribe.ExpireTime)
	expiredDays := int(time.Since(expireTime).Hours() / 24)

	if expiredDays > expiredGroup.ExpiredDaysLimit {
		r.log.Debugf("User subscription expired %d days, exceeds limit %d days", expiredDays, expiredGroup.ExpiredDaysLimit)
		return nil
	}

	// 3. 检查用户已使用流量是否超过限制(仅使用过期期间的流量)
	if expiredGroup.MaxTrafficGBExpired != nil && *expiredGroup.MaxTrafficGBExpired > 0 {
		expiredDownload := userSubscribe.ExpiredDownload
		expiredUpload := userSubscribe.ExpiredUpload

		usedTrafficGB := float64(expiredDownload+expiredUpload) / (1024 * 1024 * 1024)
		if usedTrafficGB >= float64(*expiredGroup.MaxTrafficGBExpired) {
			r.log.Debugf("User expired traffic %.2f GB, exceeds expired group limit %.2f GB",
				usedTrafficGB, float64(*expiredGroup.MaxTrafficGBExpired))
			return nil
		}
	}

	// 4. 查询过期节点组的节点
	nodes, err := r.data.db.ProxyNode.Query().
		Where(
			proxynode.EnabledEQ(true),
		).
		Where(func(s *sql.Selector) {
			s.Where(sql.ExprP("JSON_CONTAINS("+proxynode.FieldNodeGroupIds+", ?)", fmt.Sprintf("%d", expiredGroup.ID)))
		}).
		Order(ent.Asc(proxynode.FieldSort)).
		Limit(1000).
		All(ctx)

	if err != nil {
		r.log.Errorf("Failed to query expired group nodes: %v", err)
		return r.createExpiredNodesDefault()
	}

	if len(nodes) == 0 {
		r.log.Debug("No nodes found in expired group")
		return r.createExpiredNodesDefault()
	}

	// 5. 转换为NodeInfo
	nodeInfos := make([]*subscriptionbiz.NodeInfo, 0, len(nodes))

	for _, node := range nodes {
		// 获取服务器信息
		server, err := r.data.db.ProxyServer.Query().
			Where(proxyserver.IDEQ(node.ServerID)).
			Only(ctx)

		if err != nil {
			r.log.Warnf("Failed to query server for node %d: %v", node.ID, err)
			continue
		}

		protocols, err := servermodel.UnmarshalProtocols(server.Protocol)
		if err != nil {
			r.log.Warnf("Failed to unmarshal server protocols for server %d: %v", server.ID, err)
			continue
		}

		var matched *servermodel.Protocol
		for _, protocol := range protocols {
			if protocol != nil && protocol.Type == node.Protocol {
				matched = protocol
				break
			}
		}
		if matched == nil {
			continue
		}

		nodeInfo := &subscriptionbiz.NodeInfo{
			ID:                      int64(node.ID),
			Sort:                    node.Sort,
			Name:                    node.Name,
			Server:                  node.Address,
			Port:                    node.Port,
			Type:                    node.Protocol,
			Tags:                    tool.StringToStringSlice(node.Tags),
			NodeGroupID:             expiredGroup.ID,
			Security:                matched.Security,
			SNI:                     matched.SNI,
			AllowInsecure:           matched.AllowInsecure,
			Fingerprint:             matched.Fingerprint,
			Method:                  matched.Cipher,
			ServerKey:               matched.ServerKey,
			Flow:                    matched.Flow,
			Transport:               matched.Transport,
			Host:                    matched.Host,
			Path:                    matched.Path,
			ServiceName:             matched.ServiceName,
			RealityServerAddr:       matched.RealityServerAddr,
			RealityServerPort:       int(matched.RealityServerPort),
			RealityPrivateKey:       matched.RealityPrivateKey,
			RealityPublicKey:        matched.RealityPublicKey,
			RealityShortId:          matched.RealityShortId,
			HopPorts:                matched.HopPorts,
			HopInterval:             int(matched.HopInterval),
			ObfsPassword:            matched.ObfsPassword,
			UpMbps:                  int(matched.UpMbps),
			DownMbps:                int(matched.DownMbps),
			DisableSNI:              matched.DisableSNI,
			ReduceRtt:               matched.ReduceRtt,
			UDPRelayMode:            matched.UDPRelayMode,
			CongestionController:    matched.CongestionController,
			PaddingScheme:           matched.PaddingScheme,
			Multiplex:               matched.Multiplex,
			XhttpMode:               matched.XhttpMode,
			XhttpExtra:              matched.XhttpExtra,
			Encryption:              matched.Encryption,
			EncryptionMode:          matched.EncryptionMode,
			EncryptionRtt:           matched.EncryptionRtt,
			EncryptionTicket:        matched.EncryptionTicket,
			EncryptionServerPadding: matched.EncryptionServerPadding,
			EncryptionPrivateKey:    matched.EncryptionPrivateKey,
			EncryptionClientPadding: matched.EncryptionClientPadding,
			EncryptionPassword:      matched.EncryptionPassword,
			Ratio:                   matched.Ratio,
			CertMode:                matched.CertMode,
			CertDNSProvider:         matched.CertDNSProvider,
			CertDNSEnv:              matched.CertDNSEnv,
		}

		nodeInfos = append(nodeInfos, nodeInfo)
	}

	r.log.Infof("Returned %d nodes from expired group for user %d (expired %d days)",
		len(nodeInfos), userSubscribe.UserID, expiredDays)

	return nodeInfos
}

// createExpiredNodesDefault 创建默认的过期提示节点
func (r *publicSubscriptionRepo) createExpiredNodesDefault() []*subscriptionbiz.NodeInfo {
	host := r.getFirstHostLine()

	return []*subscriptionbiz.NodeInfo{
		{
			Name:   "Subscribe Expired",
			Server: "127.0.0.1",
			Port:   18080,
			Type:   "shadowsocks",
			Tags:   []string{},
			Sort:   1,
			Method: "aes-256-gcm",
		},
		{
			Name:   host,
			Server: "127.0.0.1",
			Port:   18080,
			Type:   "shadowsocks",
			Tags:   []string{},
			Sort:   2,
			Method: "aes-256-gcm",
		},
	}
}
