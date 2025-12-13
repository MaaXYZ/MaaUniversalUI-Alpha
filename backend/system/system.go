package system

import (
	"context"
	"sync"
)

var (
	srvInst *service
	srvOnce sync.Once
)

func System() *service {
	srvOnce.Do(func() {
		srvInst = &service{}
	})
	return srvInst
}

func Startup(ctx context.Context) {
	srv := System()
	srv.ctx = ctx
}
