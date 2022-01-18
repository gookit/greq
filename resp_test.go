package hreq_test

import (
	"fmt"
	"testing"

	"github.com/gookit/hreq"
	"github.com/stretchr/testify/assert"
)

func TestResponse_String(t *testing.T) {
	resp, err := hreq.New(testBaseURL).
		UserAgent("custom-client/1.0").
		Send("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.False(t, resp.IsFail())

	fmt.Print(resp.String())
}
