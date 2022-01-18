package hreq_test

import (
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/jsonutil"
	"github.com/gookit/goutil/netutil/httpreq"
	"github.com/gookit/hreq"
	"github.com/stretchr/testify/assert"
)

func init() {
	hreq.BaseURL(testBaseURL)
}

func TestGet(t *testing.T) {
	resp, err := hreq.Std().
		JSONType().
		Get("/get")

	assert.NoError(t, err)
	sc := resp.StatusCode
	assert.True(t, httpreq.IsOK(sc))
	assert.True(t, httpreq.IsSuccessful(sc))

	retMp := make(map[string]interface{})
	err = jsonutil.DecodeReader(resp.Body, &retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}
