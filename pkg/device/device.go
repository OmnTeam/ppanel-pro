package device

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gorilla/websocket"
)

type Operator int

const (
	MaxDevices Operator = iota
	Admin
	SubscribeUpdate = "subscribe_update"
)

// Device represents a device structure
type Device struct {
	Session      string
	DeviceID     string
	Conn         *websocket.Conn
	CreatedAt    time.Time
	LastPingTime time.Time
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// DeviceManager manages devices
type DeviceManager struct {
	userDevices      sync.Map // userID -> []*Device
	totalOnline      int32    // total online devices
	userMutexes      sync.Map // userID level locks
	heartbeatTimeout int      // heartbeat timeout (seconds)
	checkInterval    int      // heartbeat check interval (seconds)
	logger           *log.Helper

	// event callbacks
	OnDeviceOnline  func(userID int64, deviceID, session string)
	OnDeviceOffline func(userID int64, deviceID, session string, createAt time.Time)
	OnDeviceKicked  func(userID int64, deviceID, session string, operator Operator)
	OnMessage       func(userID int64, deviceID, session string, message string)
}

// Get user-level mutex
func (dm *DeviceManager) getUserMutex(userID int64) *sync.Mutex {
	mu, _ := dm.userMutexes.LoadOrStore(userID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// Listen to WebSocket data
func (dm *DeviceManager) listenToDevice(userID int64, device *Device) {
	defer func() {
		dm.removeDevice(userID, device.DeviceID) // remove device when disconnected
	}()

	for {
		_, msg, err := device.Conn.ReadMessage()
		if err != nil {
			dm.logger.Infof("Device %s (User %d) disconnected: %v", device.DeviceID, userID, err)
			break
		}

		message := string(msg)
		if message == "ping" || message == "heartbeat" {
			dm.UpdateHeartbeat(userID, device.DeviceID)
			continue
		}

		// Trigger message callback
		if dm.OnMessage != nil {
			go dm.OnMessage(userID, device.DeviceID, device.Session, message)
		}
	}
}

// UpdateHeartbeat updates device heartbeat
func (dm *DeviceManager) UpdateHeartbeat(userID int64, deviceID string) {
	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	if val, ok := dm.userDevices.Load(userID); ok {
		devices := val.([]*Device)
		for _, d := range devices {
			if d.DeviceID == deviceID {
				d.LastPingTime = time.Now()
				if err := d.Conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
					dm.logger.Infof("✅ Heartbeat updated: Device %s (User %d) err: %s", deviceID, userID, err.Error())
				}
				break
			}
		}
	}
}

// AddDevice **Add: Device connects WebSocket and is added to the manager**
func (dm *DeviceManager) AddDevice(w http.ResponseWriter, r *http.Request, session string, userID int64, deviceID string, maxDevices int) error {
	// **Upgrade WebSocket connection**
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		dm.logger.Infof("WebSocket upgrade failed: %v", err)
		return err
	}

	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	newDevice := &Device{
		Session:      session,
		DeviceID:     deviceID,
		Conn:         conn,
		CreatedAt:    time.Now(),
		LastPingTime: time.Now(),
	}

	//不限制设备数量
	if maxDevices < 1 {
		maxDevices = 99
	}

	// Get user's device list
	var restConnection bool
	var devices []*Device
	if val, ok := dm.userDevices.Load(userID); ok {
		devices = val.([]*Device)
		var tempDevice []*Device
		for _, d := range devices {
			if d.DeviceID == deviceID {
				restConnection = true
			} else {
				tempDevice = append(tempDevice, d)
			}
		}
		devices = tempDevice
	}

	// **If exceeding the limit, kick out the earliest device**
	if !restConnection && len(devices) >= maxDevices {
		oldestDevice := devices[0]
		devices = devices[1:]

		if dm.OnDeviceKicked != nil {
			done := make(chan struct{})
			go func() {
				defer close(done)
				dm.OnDeviceKicked(userID, oldestDevice.DeviceID, oldestDevice.Session, MaxDevices)
			}()
			<-done // block and wait for callback to complete
		}
		oldestDevice.Conn.Close()
		atomic.AddInt32(&dm.totalOnline, -1)
	}

	// Add new device
	devices = append(devices, newDevice)
	dm.userDevices.Store(userID, devices)
	atomic.AddInt32(&dm.totalOnline, 1)

	// Trigger online event
	if dm.OnDeviceOnline != nil {
		go dm.OnDeviceOnline(userID, deviceID, session)
	}

	// Start listening
	go dm.listenToDevice(userID, newDevice)

	return nil
}

// removeDevice removes a device
func (dm *DeviceManager) removeDevice(userID int64, deviceID string) {
	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	if val, ok := dm.userDevices.Load(userID); ok {
		devices := val.([]*Device)
		for i, d := range devices {
			if d.DeviceID == deviceID {
				devices = append(devices[:i], devices[i+1:]...)
				d.Conn.Close()
				atomic.AddInt32(&dm.totalOnline, -1)

				if dm.OnDeviceOffline != nil {
					go dm.OnDeviceOffline(userID, deviceID, d.Session, d.CreatedAt)
				}
				break
			}
		}

		if len(devices) == 0 {
			dm.userDevices.Delete(userID)
		} else {
			dm.userDevices.Store(userID, devices)
		}
	}
}

// KickDevice kicks a device (supports individual device or entire user)
func (dm *DeviceManager) KickDevice(userID int64, deviceID string) {
	mu := dm.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// Get user's device list
	val, ok := dm.userDevices.Load(userID)
	if !ok {
		dm.logger.Infof("⚠️ User %d has no online devices, unable to kick out", userID)
		return
	}

	devices := val.([]*Device)
	var activeDevices []*Device

	for _, d := range devices {
		if deviceID == "" || d.DeviceID == deviceID {
			// Trigger kick event callback
			if dm.OnDeviceKicked != nil {
				done := make(chan struct{})
				go func() {
					defer close(done)
					dm.OnDeviceKicked(userID, d.DeviceID, d.Session, Admin)
				}()
				<-done // block and wait for callback to complete
			}
			// Close WebSocket connection
			d.Conn.Close()
			atomic.AddInt32(&dm.totalOnline, -1)
			dm.logger.Infof("❌ Device %s (User %d) kicked out", d.DeviceID, userID)
		} else {
			activeDevices = append(activeDevices, d)
		}
	}

	// Update user's device mapping
	if len(activeDevices) == 0 {
		dm.userDevices.Delete(userID)
	} else {
		dm.userDevices.Store(userID, activeDevices)
	}
}

// StartHeartbeatCheck periodically checks for heartbeat timeout devices
func (dm *DeviceManager) StartHeartbeatCheck() {
	ticker := time.NewTicker(time.Duration(dm.checkInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		dm.userDevices.Range(func(userID, val interface{}) bool {
			uid := userID.(int64)
			devices := val.([]*Device)

			mu := dm.getUserMutex(uid)
			mu.Lock()
			defer mu.Unlock()

			var activeDevices []*Device
			for _, d := range devices {
				if now.Sub(d.LastPingTime) > time.Duration(dm.heartbeatTimeout)*time.Second {
					dm.logger.Infof("⚠️ Device %s (User %d) heartbeat timeout, removed", d.DeviceID, uid)
					d.Conn.Close()
					atomic.AddInt32(&dm.totalOnline, -1)

					if dm.OnDeviceOffline != nil {
						go dm.OnDeviceOffline(uid, d.DeviceID, d.Session, d.CreatedAt)
					}
				} else {
					activeDevices = append(activeDevices, d)
				}
			}

			if len(activeDevices) == 0 {
				dm.userDevices.Delete(uid)
			} else {
				dm.userDevices.Store(uid, activeDevices)
			}
			return true
		})
		//zap.S().Infof("Total online devices: %d\n", dm.totalOnline)
	}
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(logger log.Logger, heartbeatTimeout, checkInterval int) *DeviceManager {
	dm := &DeviceManager{
		heartbeatTimeout: heartbeatTimeout,
		checkInterval:    checkInterval,
		logger:           log.NewHelper(logger),
	}
	go dm.StartHeartbeatCheck()
	return dm
}

// SendToDevice sends a message to a specific device
func (dm *DeviceManager) SendToDevice(userID int64, deviceID string, message string) error {
	if val, ok := dm.userDevices.Load(userID); ok {
		devices := val.([]*Device)
		if deviceID == "" {
			for _, d := range devices {
				err := d.Conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					return err
				}
				continue
			}
		} else {
			for _, d := range devices {
				if d.DeviceID == deviceID {
					return d.Conn.WriteMessage(websocket.TextMessage, []byte(message))
				}
			}
		}

	}
	return fmt.Errorf("device %s (User %d) is offline", deviceID, userID)
}

// Broadcast sends a message to all devices
func (dm *DeviceManager) Broadcast(message string) {
	go func(message string) {
		dm.userDevices.Range(func(_, val interface{}) bool {
			devices := val.([]*Device)
			for _, d := range devices {
				_ = d.Conn.WriteMessage(websocket.TextMessage, []byte(message))
			}
			return true
		})
	}(message)

}

// Gracefully shut down all WebSocket connections
func (dm *DeviceManager) Shutdown(ctx context.Context) {
	<-ctx.Done()
	dm.logger.Info("🔴 Shutting down all WebSocket connections...")

	dm.userDevices.Range(func(userID, val interface{}) bool {
		uid := userID.(int64)
		devices := val.([]*Device)

		for _, d := range devices {
			d.Conn.Close()
			dm.logger.Infof("✅ Closed device %s (User %d)", d.DeviceID, uid)
		}
		dm.userDevices.Delete(uid)
		return true
	})
}

// UserDeviceInfo represents user device information for API responses
type UserDeviceInfo struct {
	ID         int64
	IP         string
	Identifier string
	UserAgent  string
	Online     bool
	Enabled    bool
	CreatedAt  int64
	UpdatedAt  int64
}

// WeeklyStat represents weekly device usage statistics
type WeeklyStat struct {
	Day     int32
	DayName string
	Hours   float64
}

// ConnectionRecords represents device connection statistics
type ConnectionRecords struct {
	CurrentContinuousDays   int64
	HistoryContinuousDays   int64
	LongestSingleConnection int64
}

// DeviceStatistics represents device usage statistics
type DeviceStatistics struct {
	WeeklyStats       []*WeeklyStat
	ConnectionRecords *ConnectionRecords
}

// GetUserDevices gets user device list for API
func (dm *DeviceManager) GetUserDevices(userID int64) ([]*UserDeviceInfo, error) {
	var devices []*UserDeviceInfo

	if val, ok := dm.userDevices.Load(userID); ok {
		deviceList := val.([]*Device)
		for _, d := range deviceList {
			devices = append(devices, &UserDeviceInfo{
				ID:         0, // Will be set by database layer
				IP:         "", // Will be set by database layer
				Identifier: d.DeviceID,
				UserAgent:  "", // Will be set by database layer
				Online:     true,
				Enabled:    true,
				CreatedAt:  d.CreatedAt.UnixMilli(),
				UpdatedAt:  d.LastPingTime.UnixMilli(),
			})
		}
	}

	return devices, nil
}

// RemoveDevice removes a device for API
func (dm *DeviceManager) RemoveDevice(userID int64, deviceID int64) error {
	// Convert deviceID to string and remove the device
	dm.removeDevice(userID, fmt.Sprintf("%d", deviceID))
	return nil
}

// GetUserDeviceStatistics gets device usage statistics for API
func (dm *DeviceManager) GetUserDeviceStatistics(userID int64) (*DeviceStatistics, error) {
	// Generate sample statistics
	weeklyStats := []*WeeklyStat{
		{Day: 1, DayName: "Monday", Hours: 2.5},
		{Day: 2, DayName: "Tuesday", Hours: 3.2},
		{Day: 3, DayName: "Wednesday", Hours: 1.8},
		{Day: 4, DayName: "Thursday", Hours: 4.1},
		{Day: 5, DayName: "Friday", Hours: 2.9},
		{Day: 6, DayName: "Saturday", Hours: 5.2},
		{Day: 7, DayName: "Sunday", Hours: 3.7},
	}

	connectionRecords := &ConnectionRecords{
		CurrentContinuousDays:   3,
		HistoryContinuousDays:   7,
		LongestSingleConnection: 86400, // 24 hours in seconds
	}

	return &DeviceStatistics{
		WeeklyStats:       weeklyStats,
		ConnectionRecords: connectionRecords,
	}, nil
}