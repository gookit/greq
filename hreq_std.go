package hreq

// std instance
var std = New()

// Std instance
func Std() *HReq {
	return std
}

// BaseURL set base URL for request
func BaseURL(baseURL string) *HReq {
	return std.BaseURL(baseURL)
}
