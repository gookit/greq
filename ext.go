package greq

import (
	"net/http"
	gourl "net/url"

	"github.com/gookit/goutil/arrutil"
)

// SetHeaders sets the key, value pairs from the given http.Header to the
// request. Values for existing keys are overwritten.
//
// TODO replace to goutil/netutil/httpreq.SetHeaders
func SetHeaders(req *http.Request, headers ...http.Header) {
	for _, header := range headers {
		for key, values := range header {
			req.Header[key] = values
		}
	}
}

// SetHeaderMap to reqeust instance.
//
// TODO replace to goutil/netutil/httpreq.SetHeaderMap
func SetHeaderMap(req *http.Request, headerMap map[string]string) {
	for k, v := range headerMap {
		req.Header.Set(k, v)
	}
}

// MergeURLValues merge url.Values by overwrite.
//
// values support: url.Values, map[string]string, map[string][]string
//
// TODO replace to goutil/netutil/httpreq.MergeURLValues
func MergeURLValues(uv gourl.Values, values ...any) gourl.Values {
	if uv == nil {
		uv = make(gourl.Values)
	}

	for _, v := range values {
		switch tv := v.(type) {
		case gourl.Values:
			for k, vs := range tv {
				uv[k] = vs
			}
		case map[string]any:
			for k, v := range tv {
				uv[k] = arrutil.AnyToStrings(v)
			}
		case map[string]string:
			for k, v := range tv {
				uv[k] = []string{v}
			}
		case map[string][]string:
			for k, vs := range tv {
				uv[k] = vs
			}
		}
	}

	return uv
}
