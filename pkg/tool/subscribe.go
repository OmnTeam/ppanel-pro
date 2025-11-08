package tool

import (
	"time"
)

// CalculateNextResetTime 计算下次重置时间
// 与原项目 logic 完全一致
func CalculateNextResetTime(expireTime int64, resetCycle int32) int64 {
	resetTime := time.UnixMilli(expireTime)
	now := time.Now()

	switch resetCycle {
	case 0:
		return 0 // 不重置
	case 1:
		// 按月重置 - 下个月1号
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).UnixMilli()
	case 2:
		// 按天重置 - 每月同一天
		if resetTime.Day() > now.Day() {
			return time.Date(now.Year(), now.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		} else {
			return time.Date(now.Year(), now.Month()+1, resetTime.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		}
	case 3:
		// 按年重置 - 每年同月同日
		targetTime := time.Date(now.Year(), resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		if targetTime.Before(now) {
			targetTime = time.Date(now.Year()+1, resetTime.Month(), resetTime.Day(), 0, 0, 0, 0, now.Location())
		}
		return targetTime.UnixMilli()
	default:
		return 0
	}
}