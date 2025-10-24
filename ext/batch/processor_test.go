package batch_test

import (
	"context"
	"testing"

	"github.com/gookit/goutil/testutil"
	"github.com/gookit/goutil/testutil/assert"
	"github.com/gookit/greq"
	"github.com/gookit/greq/ext/batch"
)

func TestAddPutAndAddDelete(t *testing.T) {
	bp := batch.NewProcessor(
		batch.WithClient(greq.New()),
		batch.WithContext(context.Background()),
	)
	bp.AddPut("put_data", testApiURL + "/put", map[string]string{"test": "data", "type": "put"})
	bp.AddDelete("delete_req", testApiURL + "/delete")

	results := bp.ExecuteAll()
	assert.Len(t, results, 2)

	// check put result
	putResult := results["put_data"]
	assert.NotEmpty(t, putResult)
	assert.Nil(t, putResult.Error)
	assert.True(t, putResult.Response.IsOK())
	respData := testutil.ParseRespToReply(putResult.Response.Response)
	assert.StrContains(t, respData.Body, "type=put")
	assert.StrContains(t, respData.Body, "test=data")

	// check delete result
	deleteResult := results["delete_req"]
	assert.NotEmpty(t, deleteResult)
	assert.Nil(t, deleteResult.Error)
	assert.True(t, deleteResult.Response.IsOK())
	respData = testutil.ParseRespToReply(deleteResult.Response.Response)
	assert.Eq(t, "DELETE", respData.Method)
	assert.StrContains(t, respData.URL, "/delete")
}
