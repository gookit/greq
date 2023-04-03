package greq

import (
	"context"
	"time"
)

func ensureOpt(opt *ReqOption) *ReqOption {
	if opt == nil {
		opt = &ReqOption{}
	}
	if opt.Context == nil {
		opt.Context = context.Background()
	}

	if opt.Timeout > 0 {
		opt.Context, opt.TCancelFunc = context.WithTimeout(
			opt.Context,
			time.Duration(opt.Timeout)*time.Millisecond,
		)
	}
	return opt
}
