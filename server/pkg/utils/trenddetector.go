// From https://github.com/livekit/livekit/blob/master/pkg/sfu/streamallocator/trenddetector.go
package utils

import (
	"fmt"
	"time"
)

// ------------------------------------------------

type TrendDirection int

const (
	TrendDirectionNeutral TrendDirection = iota
	TrendDirectionUpward
	TrendDirectionDownward
)

func (t TrendDirection) String() string {
	switch t {
	case TrendDirectionNeutral:
		return "NEUTRAL"
	case TrendDirectionUpward:
		return "UPWARD"
	case TrendDirectionDownward:
		return "DOWNWARD"
	default:
		return fmt.Sprintf("%d", int(t))
	}
}

// ------------------------------------------------

type TrendDetectorParams struct {
	RequiredSamples        int
	DownwardTrendThreshold float64
	CollapseValues         bool
}

type TrendDetector struct {
	params TrendDetectorParams

	startTime    time.Time
	numSamples   int
	values       []int64
	lowestValue  int64
	highestValue int64

	direction TrendDirection
}

func NewTrendDetector(params TrendDetectorParams) *TrendDetector {
	return &TrendDetector{
		params:    params,
		startTime: time.Now(),
		direction: TrendDirectionNeutral,
	}
}

func (t *TrendDetector) Seed(value int64) {
	if len(t.values) != 0 {
		return
	}

	t.values = append(t.values, value)
}

func (t *TrendDetector) AddValue(value int64) {
	t.numSamples++
	if t.lowestValue == 0 || value < t.lowestValue {
		t.lowestValue = value
	}
	if value > t.highestValue {
		t.highestValue = value
	}

	// ignore duplicate values
	if t.params.CollapseValues && len(t.values) != 0 && t.values[len(t.values)-1] == value {
		return
	}

	if len(t.values) == t.params.RequiredSamples {
		t.values = t.values[1:]
	}
	t.values = append(t.values, value)

	t.updateDirection()
}

func (t *TrendDetector) GetLowest() int64 {
	return t.lowestValue
}

func (t *TrendDetector) GetHighest() int64 {
	return t.highestValue
}

func (t *TrendDetector) GetValues() []int64 {
	return t.values
}

func (t *TrendDetector) GetDirection() TrendDirection {
	return t.direction
}

func (t *TrendDetector) ToString() string {
	now := time.Now()
	elapsed := now.Sub(t.startTime).Seconds()
	str := fmt.Sprintf("t: %+v|%+v|%.2fs", t.startTime.Format(time.UnixDate), now.Format(time.UnixDate), elapsed)
	str += fmt.Sprintf(", v: %d|%d|%d|%+v|%.2f", t.numSamples, t.lowestValue, t.highestValue, t.values, kendallsTau(t.values))
	return str
}

func (t *TrendDetector) updateDirection() {
	if len(t.values) < t.params.RequiredSamples {
		t.direction = TrendDirectionNeutral
		return
	}

	// using Kendall's Tau to find trend
	kt := kendallsTau(t.values)

	t.direction = TrendDirectionNeutral
	switch {
	case kt > 0:
		t.direction = TrendDirectionUpward
	case kt < t.params.DownwardTrendThreshold:
		t.direction = TrendDirectionDownward
	}
}

// ------------------------------------------------

func kendallsTau(values []int64) float64 {
	concordantPairs := 0
	discordantPairs := 0

	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] < values[j] {
				concordantPairs++
			} else if values[i] > values[j] {
				discordantPairs++
			}
		}
	}

	if (concordantPairs + discordantPairs) == 0 {
		return 0.0
	}

	return (float64(concordantPairs) - float64(discordantPairs)) / (float64(concordantPairs) + float64(discordantPairs))
}
