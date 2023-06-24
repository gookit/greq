package greq_test

import (
	"fmt"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
)

func TestResponse_String(t *testing.T) {
	resp, err := greq.New(testBaseURL).
		UserAgent("custom-cli/1.0").
		GetDo("/get")

	assert.NoErr(t, err)
	assert.True(t, resp.IsOK())
	assert.False(t, resp.IsFail())

	fmt.Print(resp.String())
}
