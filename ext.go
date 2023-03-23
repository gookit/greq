package greq

func ensureOpt(opt *ReqOption) *ReqOption {
	if opt == nil {
		opt = &ReqOption{}
	}
	return opt
}
