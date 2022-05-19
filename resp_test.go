package hireq_test

import (
	"fmt"
	"testing"

	"github.com/gookit/hireq"
	"github.com/stretchr/testify/assert"
)

func TestResponse_String(t *testing.T) {
	resp, err := hireq.New(testBaseURL).
		UserAgent("custom-client/1.0").
		GetDo("/get")

	assert.NoError(t, err)
	assert.True(t, resp.IsOK())
	assert.False(t, resp.IsFail())

	fmt.Print(resp.String())
}
