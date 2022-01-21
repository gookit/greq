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

func TestGet(t *testing.T) {
	resp, err := hreq.Get("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestGet_with_QueryParams(t *testing.T) {
	resp, err := hreq.Std().
		JSONType().
		QueryParams(map[string]string{
			"name": "inhere",
		}).
		Get("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())
	assert.True(t, resp.IsJSONType())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestPost(t *testing.T) {
	resp, err := hreq.Post("/post")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestPut(t *testing.T) {
	resp, err := hreq.Put("/put")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestPatch(t *testing.T) {
	resp, err := hreq.Patch("/patch")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestDelete(t *testing.T) {
	resp, err := hreq.Delete("/delete")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())

	retMp := make(map[string]interface{})
	err = resp.Decode(&retMp)
	assert.NoError(t, err)
	dump.P(retMp)
}

func TestHead(t *testing.T) {
	resp, err := hreq.Head("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())
	assert.False(t, resp.IsEmptyBody())

	assert.Empty(t, resp.BodyString())
}

func TestOptions(t *testing.T) {
	resp, err := hreq.Options("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsEmptyBody())
	assert.True(t, resp.IsSuccessful())

	assert.Empty(t, resp.BodyString())
	assert.NotEmpty(t, resp.HeaderString())
}

func TestTrace(t *testing.T) {
	resp, err := hreq.Trace("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
	// assert.True(t, resp.IsNoBody())
}

func TestConnect(t *testing.T) {
	resp, err := hreq.Connect("/")
	fmt.Println(resp.String())

	assert.NoError(t, err)
}
