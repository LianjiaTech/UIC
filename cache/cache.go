package cache

import (
	"github.com/open-falcon/fe/g"
	"github.com/toolkits/cache"
	"time"
)

func InitCache() {
	cfg := g.Config()
	if !cfg.Cache.Enabled {
		return
	}

	cache.InitCache(
		cfg.Cache.Redis,
		cfg.Cache.Idle,
		cfg.Cache.Max,
		time.Duration(cfg.Cache.Timeout.Conn)*time.Millisecond,
		time.Duration(cfg.Cache.Timeout.Read)*time.Millisecond,
		time.Duration(cfg.Cache.Timeout.Write)*time.Millisecond,
		time.Hour,
	)
}
