package greq_test

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/dump"
	"github.com/gookit/goutil/testutil"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
)

func TestREST_methodDo(t *testing.T) {
	t.Run("GET", func(t *testing.T) {
		resp, err := greq.GetDo("/get")

		assert.NoErr(t, err)
		assert.True(t, resp.IsOK())
		assert.True(t, resp.IsSuccessful())

		retMp := make(map[string]any)
		err = resp.Decode(&retMp)
		assert.NoErr(t, err)
		dump.P(retMp)
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := greq.PostDo("/post", `{"name": "inhere"}`)

		assert.NoErr(t, err)
		assert.True(t, resp.IsOK())
		assert.True(t, resp.IsSuccessful())

		retMp := make(map[string]any)
		err = resp.Decode(&retMp)
		assert.NoErr(t, err)
		dump.P(retMp)
	})

	t.Run("PUT", func(t *testing.T) {
		resp, err := greq.PutDo(testBaseURL+"/put", `{"name": "inhere"}`)

		assert.NoErr(t, err)
		assert.True(t, resp.IsOK())
		assert.True(t, resp.IsSuccessful())

		retMp := make(map[string]any)
		err = resp.Decode(&retMp)
		assert.NoErr(t, err)
		dump.P(retMp)
	})

	t.Run("PATCH", func(t *testing.T) {
		resp, err := greq.PatchDo(testBaseURL+"/patch", "hello")

		assert.NoErr(t, err)
		assert.True(t, resp.IsOK())
		assert.True(t, resp.IsSuccessful())

		rr := &testutil.EchoReply{}
		err = resp.Decode(&rr)
		assert.NoErr(t, err)
		dump.P(rr)
	})

	t.Run("DELETE", func(t *testing.T) {
		resp, err := greq.DeleteDo(testBaseURL + "/delete")

		assert.NoErr(t, err)
		assert.True(t, resp.IsOK())
		assert.True(t, resp.IsSuccessful())

		retMp := make(map[string]any)
		err = resp.Decode(&retMp)
		assert.NoErr(t, err)
		dump.P(retMp)
	})
}

func TestGetDo_with_QueryParams(t *testing.T) {
	resp, err := greq.Std().
		JSONType().
		QueryParams(map[string]string{
			"name": "inhere",
		}).
		GetDo("/get")

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())
	assert.True(t, resp.IsSuccessful())
	assert.True(t, resp.IsJSONType())

	retMp := make(map[string]any)
	err = resp.Decode(&retMp)
	assert.NoErr(t, err)
	dump.P(retMp)
}

func TestOther_methodDo(t *testing.T) {
	t.Run("Head", func(t *testing.T) {
		resp, err := greq.Reset().HeadDo(testBaseURL + "/")
		fmt.Println(resp.String())

		assert.NoErr(t, err)
		assert.True(t, resp.IsOK())
		assert.True(t, resp.IsSuccessful())
		assert.False(t, resp.IsEmptyBody())

		assert.Empty(t, resp.BodyString())
	})

	t.Run("Options", func(t *testing.T) {
		resp, err := greq.OptionsDo(testBaseURL + "/")
		fmt.Println(resp.String())

		assert.NoErr(t, err)
		assert.True(t, resp.IsOK())
		assert.True(t, resp.IsSuccessful())
		assert.False(t, resp.IsEmptyBody())

		assert.Empty(t, resp.BodyString())
		assert.NotEmpty(t, resp.HeaderString())
	})

	t.Run("Trace", func(t *testing.T) {
		resp, err := greq.TraceDo(testBaseURL + "/")
		fmt.Println(resp.String())

		assert.NoErr(t, err)
		// assert.True(t, resp.IsNoBody())
	})

	t.Run("Connect", func(t *testing.T) {
		resp, err := greq.ConnectDo(testBaseURL + "/")
		fmt.Println(resp.String())

		assert.NoErr(t, err)
	})
}
