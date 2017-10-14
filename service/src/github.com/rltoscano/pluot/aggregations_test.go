package pluot

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestInvalidJson(t *testing.T) {
	c, done := aeContext(t)
	defer done()
	r := createReq(t, "invalidjson")
	if _, err := computeAggregation(c, r, nil); err == nil {
		t.Fatal("Expected error, but was nil")
	}
}

func TestOneDayAgg(t *testing.T) {
	c, done := aeContext(t)
	defer done()
	postDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", "Tue, 10 Oct 2017 00:00:00 GMT")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	txn := Txn{PostDate: postDate, Category: 3, Amount: 1000}
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 1, nil), &txn); err != nil {
		t.Fatal(err)
	}
	d, _ := time.ParseDuration("500ms")
	time.Sleep(d)
	r := createReq(t, `{
    "start": "Tue, 10 Oct 2017 00:00:00 GMT",
    "end": "Tue, 11 Oct 2017 00:00:00 GMT"
  }`)
	resp, err := computeAggregation(c, r, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.(ComputeAggregationResponse).Totals[3] != 1000 {
		t.Fatalf("Wanted resp.Totals[3] to be 1000, but was %v", resp.(ComputeAggregationResponse).Totals[3])
	}
}

func aeContext(t *testing.T) (context.Context, func()) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	return c, done
}

func createReq(t *testing.T, body string) *http.Request {
	r, err := http.NewRequest("method", "url", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
