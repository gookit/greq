package batch_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gookit/goutil/testutil"
	"github.com/gookit/greq/ext/batch"
)

var testApiURL string

func TestMain(m *testing.M) {
	// create mock server
	s := testutil.NewEchoServer()
	defer s.Close()
	testApiURL = s.HTTPHost()
	s.PrintHttpHost()

	// do testing
	m.Run()
}

func Example() {
	// Execute all requests
	bp := batch.NewProcessor(
		batch.WithMaxConcurrency(5),
		batch.WithBatchTimeout(10*time.Second),
	)

	bp.AddGet("req1", "https://api1.example.com/data")
	bp.AddPost("req2", "https://api2.example.com/submit", map[string]string{"key": "value"})

	results := bp.ExecuteAll()
	fmt.Println("Results: ", len(results))

	// Execute any (first success)
	bp2 := batch.NewProcessor()
	bp2.AddGet("mirror1", testApiURL + "/file1")
	bp2.AddGet("mirror2", "https://mirror2.example.com/file")
	bp2.AddGet("mirror3", "https://mirror3.example.com/file")

	result := bp2.ExecuteAny()
	fmt.Println("First successful result: ", result.ID)

	// Convenience functions
	urls := []string{"https://api1.com", "https://api2.com", "https://api3.com"}
	allResults := batch.GetAll(urls)
	fmt.Println("All results: ", len(allResults))

	firstResult := batch.GetAny(urls)
	fmt.Println("First successful result: ", firstResult.ID)
}
