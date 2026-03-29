package deduction

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

const (
	UnitTimeNoLimit = "NoLimit"
	UnitTimeYear    = "Year"
	UnitTimeMonth   = "Month"
	UnitTimeDay     = "Day"
	UnitTimeHour    = "Hour"
	UnitTimeMinute  = "Minute"

	ResetCycleNone    = 0
	ResetCycle1st     = 1
	ResetCycleMonthly = 2
	ResetCycleYear    = 3

	maxInt64 = math.MaxInt64
	minInt64 = math.MinInt64
)

var (
	ErrInvalidQuantity       = errors.New("order quantity cannot be zero or negative")
	ErrInvalidAmount         = errors.New("order amount cannot be negative")
	ErrInvalidTraffic        = errors.New("traffic values cannot be negative")
	ErrInvalidTimeRange      = errors.New("expire time must be after start time")
	ErrInvalidUnitTime       = errors.New("invalid unit time")
	ErrInvalidDeductionRatio = errors.New("deduction ratio must be between 0 and 100")
	ErrOverflow              = errors.New("calculation overflow")
)

type Subscribe struct {
	StartTime      time.Time
	ExpireTime     time.Time
	Traffic        int64
	Download       int64
	Upload         int64
	UnitTime       string
	UnitPrice      int64
	ResetCycle     int64
	DeductionRatio int64
}

type Order struct {
	Amount   int64
	Quantity int64
}

func (s *Subscribe) Validate() error {
	if s.Traffic < 0 || s.Download < 0 || s.Upload < 0 {
		return ErrInvalidTraffic
	}
	if s.Download+s.Upload > s.Traffic {
		return fmt.Errorf("download + upload (%d) cannot exceed total traffic (%d)", s.Download+s.Upload, s.Traffic)
	}
	if !s.ExpireTime.After(s.StartTime) {
		return ErrInvalidTimeRange
	}
	if s.DeductionRatio < 0 || s.DeductionRatio > 100 {
		return ErrInvalidDeductionRatio
	}
	switch s.UnitTime {
	case UnitTimeNoLimit, UnitTimeYear, UnitTimeMonth, UnitTimeDay, UnitTimeHour, UnitTimeMinute:
		return nil
	default:
		return ErrInvalidUnitTime
	}
}

func (o *Order) Validate() error {
	if o.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if o.Amount < 0 {
		return ErrInvalidAmount
	}
	return nil
}

func CalculateRemainingAmount(sub Subscribe, order Order) (int64, error) {
	if err := sub.Validate(); err != nil {
		return 0, fmt.Errorf("invalid subscription: %w", err)
	}
	if err := order.Validate(); err != nil {
		return 0, fmt.Errorf("invalid order: %w", err)
	}
	if sub.UnitTime == UnitTimeNoLimit && sub.ResetCycle != 0 {
		return 0, nil
	}

	unitPrice, err := safeDivide(order.Amount, order.Quantity)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate unit price: %w", err)
	}
	sub.UnitPrice = unitPrice

	loc, err := time.LoadLocation(sub.StartTime.Location().String())
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	switch sub.UnitTime {
	case UnitTimeNoLimit:
		return calculateNoLimitAmount(sub, order)
	case UnitTimeYear:
		remainingYears := tool.YearDiff(now, sub.ExpireTime)
		remainingUnitTimeAmount, err := calculateRemainingUnitTimeAmount(sub)
		if err != nil {
			return 0, err
		}
		yearAmount, err := safeMultiply(int64(remainingYears), sub.UnitPrice)
		if err != nil {
			return 0, fmt.Errorf("year calculation overflow: %w", err)
		}
		return safeAdd(yearAmount, remainingUnitTimeAmount)
	case UnitTimeMonth:
		remainingMonths := tool.MonthDiff(now, sub.ExpireTime)
		remainingUnitTimeAmount, err := calculateRemainingUnitTimeAmount(sub)
		if err != nil {
			return 0, err
		}
		monthAmount, err := safeMultiply(int64(remainingMonths), sub.UnitPrice)
		if err != nil {
			return 0, fmt.Errorf("month calculation overflow: %w", err)
		}
		return safeAdd(monthAmount, remainingUnitTimeAmount)
	case UnitTimeDay:
		remainingDays := tool.DayDiff(now, sub.ExpireTime)
		remainingUnitTimeAmount, err := calculateRemainingUnitTimeAmount(sub)
		if err != nil {
			return 0, err
		}
		dayAmount, err := safeMultiply(remainingDays, sub.UnitPrice)
		if err != nil {
			return 0, fmt.Errorf("day calculation overflow: %w", err)
		}
		return safeAdd(dayAmount, remainingUnitTimeAmount)
	default:
		return 0, nil
	}
}

func calculateNoLimitAmount(sub Subscribe, order Order) (int64, error) {
	if sub.Traffic == 0 {
		return 0, nil
	}

	remainingTraffic := sub.Traffic - sub.Download - sub.Upload
	if remainingTraffic < 0 {
		remainingTraffic = 0
	}

	unitPrice := float64(order.Amount) / float64(sub.Traffic)
	result := float64(remainingTraffic) * unitPrice
	if result > float64(maxInt64) || result < float64(minInt64) {
		return 0, ErrOverflow
	}
	return int64(result), nil
}

func calculateRemainingUnitTimeAmount(sub Subscribe) (int64, error) {
	now := time.Now()
	trafficWeight, timeWeight := calculateWeights(sub.DeductionRatio)
	remainingDays, totalDays := getRemainingAndTotalDays(sub, now)
	if totalDays == 0 {
		return 0, nil
	}

	remainingTraffic := sub.Traffic - sub.Download - sub.Upload
	if remainingTraffic < 0 {
		remainingTraffic = 0
	}

	remainingTimeAmount, err := calculateProportionalAmount(sub.UnitPrice, remainingDays, totalDays)
	if err != nil {
		return 0, fmt.Errorf("time amount calculation failed: %w", err)
	}
	if sub.Traffic == 0 {
		return remainingTimeAmount, nil
	}

	remainingTrafficAmount, err := calculateProportionalAmount(sub.UnitPrice, remainingTraffic, sub.Traffic)
	if err != nil {
		return 0, fmt.Errorf("traffic amount calculation failed: %w", err)
	}
	if sub.DeductionRatio != 0 {
		return calculateWeightedAmount(sub.UnitPrice, remainingTraffic, sub.Traffic, remainingDays, totalDays, trafficWeight, timeWeight)
	}
	if remainingTimeAmount < remainingTrafficAmount {
		return remainingTimeAmount, nil
	}
	return remainingTrafficAmount, nil
}

func calculateWeights(deductionRatio int64) (float64, float64) {
	if deductionRatio == 0 {
		return 0, 0
	}
	trafficWeight := float64(deductionRatio) / 100
	return trafficWeight, 1 - trafficWeight
}

func getRemainingAndTotalDays(sub Subscribe, now time.Time) (int64, int64) {
	switch sub.ResetCycle {
	case ResetCycleNone:
		remaining := sub.ExpireTime.Sub(now).Hours() / 24
		total := sub.ExpireTime.Sub(sub.StartTime).Hours() / 24
		if remaining < 0 {
			remaining = 0
		}
		if total < 0 {
			total = 0
		}
		return int64(remaining), int64(total)
	case ResetCycle1st:
		return tool.DaysToNextMonth(now), tool.GetLastDayOfMonth(now)
	case ResetCycleMonthly:
		remaining := tool.DaysToMonthDay(now, sub.StartTime.Day()) - 1
		total := tool.DaysToMonthDay(now, sub.StartTime.Day())
		if remaining < 0 {
			remaining = 0
		}
		return remaining, total
	case ResetCycleYear:
		return tool.DaysToYearDay(now, int(sub.StartTime.Month()), sub.StartTime.Day()),
			tool.GetYearDays(now, int(sub.StartTime.Month()), sub.StartTime.Day())
	default:
		return 0, 0
	}
}

func calculateWeightedAmount(unitPrice, remainingTraffic, totalTraffic, remainingDays, totalDays int64, trafficWeight, timeWeight float64) (int64, error) {
	if totalDays == 0 || totalTraffic == 0 {
		return 0, nil
	}

	remainingTimeRatio := float64(remainingDays) / float64(totalDays)
	remainingTrafficRatio := float64(remainingTraffic) / float64(totalTraffic)
	weightedRemainingRatio := timeWeight*remainingTimeRatio + trafficWeight*remainingTrafficRatio
	result := float64(unitPrice) * weightedRemainingRatio
	if result > float64(maxInt64) || result < float64(minInt64) {
		return 0, ErrOverflow
	}
	return int64(result), nil
}

func calculateProportionalAmount(total, remaining, denominator int64) (int64, error) {
	if denominator == 0 {
		return 0, nil
	}
	result := float64(total) * (float64(remaining) / float64(denominator))
	if result > float64(maxInt64) || result < float64(minInt64) {
		return 0, ErrOverflow
	}
	return int64(result), nil
}

func safeMultiply(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a > 0 && b > 0 && a > maxInt64/b {
		return 0, ErrOverflow
	}
	if a < 0 && b < 0 && a < maxInt64/b {
		return 0, ErrOverflow
	}
	if (a > 0 && b < 0 && b < minInt64/a) || (a < 0 && b > 0 && a < minInt64/b) {
		return 0, ErrOverflow
	}
	return a * b, nil
}

func safeAdd(a, b int64) (int64, error) {
	if (b > 0 && a > maxInt64-b) || (b < 0 && a < minInt64-b) {
		return 0, ErrOverflow
	}
	return a + b, nil
}

func safeDivide(a, b int64) (int64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}
