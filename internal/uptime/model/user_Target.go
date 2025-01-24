package model

import "github.com/shuvo-paul/uptimebot/internal/uptime/monitor"

type UserTarget struct {
	UserID int
	*monitor.Target
}
