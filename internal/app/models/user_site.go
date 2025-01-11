package models

import "github.com/shuvo-paul/sitemonitor/pkg/monitor"

type UserSite struct {
	UserID int
	*monitor.Site
}
