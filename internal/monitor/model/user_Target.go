package model

import monitor "github.com/shuvo-paul/uptimebot/internal/monitor/engine"

type UserTarget struct {
	UserID int
	*monitor.Target
}
