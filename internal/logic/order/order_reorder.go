package orderlogic

import "time"

type reorderConfig struct {
	SmartEnabled         int
	TimeoutEnabled       int
	TimeoutMinutes       int
	OrderStrategy        string
	AllowLossSaleEnabled int
	MaxLossAmount        string
}

func canReorder(config reorderConfig, createdAt, now time.Time) bool {
	if config.SmartEnabled != 1 || config.TimeoutEnabled != 1 || config.TimeoutMinutes <= 0 {
		return false
	}
	return !now.After(createdAt.Add(time.Duration(config.TimeoutMinutes) * time.Minute))
}

func pollIntervalDuration(seconds int) time.Duration {
	if seconds <= 0 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}
