package hreq_test

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/hreq"
	"github.com/stretchr/testify/assert"
)

func init() {
	hreq.BaseURL(testBaseURL)
}

func TestGetDo(t *testing.T) {
	resp, err := hreq.GetDo("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestGetDo_with_QueryParams(t *testing.T) {
	resp, err := hreq.Std().
		JSONType().
		QueryParams(map[string]string{
			"name": "inhere",
		}).
		GetDo("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())
	assert.True(t, resp.IsJSONType())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestPostDo(t *testing.T) {
	resp, err := hreq.PostDo("/post")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestPutDo(t *testing.T) {
	resp, err := hreq.PutDo("/put")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestPatchDo(t *testing.T) {
	resp, err := hreq.PatchDo("/patch")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestDeleteDo(t *testing.T) {
	resp, err := hreq.DeleteDo("/delete")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestHeadDo(t *testing.T) {
	resp, err := hreq.Reset().HeadDo("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())
	assert.False(t, resp.IsEmptyBody())

	assert.Empty(t, resp.BodyString())
}

func TestOptionsDo(t *testing.T) {
	resp, err := hreq.OptionsDo("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsEmptyBody())
	assert.True(t, resp.IsSuccessful())

	assert.Empty(t, resp.BodyString())
	assert.NotEmpty(t, resp.HeaderString())
}

func TestTraceDo(t *testing.T) {
	resp, err := hreq.TraceDo("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
	// assert.True(t, resp.IsNoBody())
}

func TestConnectDo(t *testing.T) {
	resp, err := hreq.ConnectDo("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
}
