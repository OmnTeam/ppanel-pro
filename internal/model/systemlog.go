package model

import "encoding/json"

// System log type constants
const (
	TypeUserTrafficRank   int8 = 40 // Top 10 User traffic rank log
	TypeServerTrafficRank int8 = 41 // Top 10 Server traffic rank log
	TypeTrafficStat       int8 = 42 // Daily traffic statistics log
)

// UserTraffic represents user traffic data
type UserTraffic struct {
	UserID      int64 `json:"user_id"`
	SubscribeID int64 `json:"subscribe_id"`
	Upload      int64 `json:"upload"`
	Download    int64 `json:"download"`
	Total       int64 `json:"total"`
}

// ServerTraffic represents server traffic data
type ServerTraffic struct {
	ServerID int64 `json:"server_id"`
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
}

// UserTrafficRank represents user traffic ranking
type UserTrafficRank struct {
	Rank []UserTraffic `json:"rank"` // Array of user traffic, ordered by rank
}

// Unmarshal implements json.Unmarshaler for UserTrafficRank
func (u *UserTrafficRank) Unmarshal(data []byte) error {
	return json.Unmarshal(data, u)
}

// Marshal implements json.Marshaler for UserTrafficRank
func (u *UserTrafficRank) Marshal() ([]byte, error) {
	return json.Marshal(u)
}

// ServerTrafficRank represents server traffic ranking
type ServerTrafficRank struct {
	Rank []ServerTraffic `json:"rank"` // Array of server traffic, ordered by rank
}

// Unmarshal implements json.Unmarshaler for ServerTrafficRank
func (s *ServerTrafficRank) Unmarshal(data []byte) error {
	return json.Unmarshal(data, s)
}

// Marshal implements json.Marshaler for ServerTrafficRank
func (s *ServerTrafficRank) Marshal() ([]byte, error) {
	return json.Marshal(s)
}

// TrafficStat represents daily traffic statistics
type TrafficStat struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
}

// Unmarshal implements json.Unmarshaler for TrafficStat
func (t *TrafficStat) Unmarshal(data []byte) error {
	return json.Unmarshal(data, t)
}

// Marshal implements json.Marshaler for TrafficStat
func (t *TrafficStat) Marshal() ([]byte, error) {
	return json.Marshal(t)
}
