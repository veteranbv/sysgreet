package collectors

import (
	"context"
	"os/user"
)

// WtmpLastLoginCollector reports the user's previous login from the login
// history (wtmp). The current session is skipped: the banner should say
// when you were last here, not that you just arrived.
type WtmpLastLoginCollector struct{}

// NewLastLoginCollector constructs the default last-login collector.
func NewLastLoginCollector() LastLoginCollector {
	return WtmpLastLoginCollector{}
}

// CollectLastLogin implements LastLoginCollector. It returns nil, not an
// error, when history is unavailable; the banner then omits the line.
func (WtmpLastLoginCollector) CollectLastLogin(ctx context.Context) (*LastLoginInfo, error) {
	u, err := user.Current()
	if err != nil {
		recordError("last_login", err)
		return nil, nil
	}
	return previousLogin(u.Username, currentTTY())
}
