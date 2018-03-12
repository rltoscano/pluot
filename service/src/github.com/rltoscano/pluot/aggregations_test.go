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
	persistTxn(c, t, Txn{ID: 1, PostDate: postDate, Category: 3, Amount: 1000})
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

func TestCategoryFilter(t *testing.T) {
	c, done := aeContext(t)
	defer done()
	postDate, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", "Tue, 10 Oct 2017 00:00:00 GMT")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	persistTxn(c, t, Txn{ID: 1, PostDate: postDate, Category: CategoryGroceries, Amount: 3000})
	persistTxn(c, t, Txn{ID: 2, PostDate: postDate, Category: CategoryEatingOut, Amount: 2000})
	d, _ := time.ParseDuration("500ms")
	time.Sleep(d)
	// Create request that filters groceries (4).
	r := createReq(t, `{
    "start": "Tue, 10 Oct 2017 00:00:00 GMT",
    "end": "Tue, 11 Oct 2017 00:00:00 GMT",
		"categoryFilter": 4
  }`)
	resp, err := computeAggregation(c, r, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	aggResp := resp.(ComputeAggregationResponse)
	if aggResp.Totals[CategoryGroceries] != 3000 {
		t.Fatalf("Wanted aggResp.Totals[CategoryGroceries] to be 3000, but was %v", aggResp.Totals[CategoryGroceries])
	}
	if aggResp.Totals[CategoryEatingOut] != 0 {
		t.Fatalf("Wanted aggResp.Totals[CategoryEatingOut] to be 0, but was %v", aggResp.Totals[CategoryEatingOut])
	}
	if len(aggResp.Months) != 1 {
		t.Fatalf("Wanted len(aggResp.Months) to be 1, but was %v", len(aggResp.Months))
	}
	if aggResp.Months[0].Expense != -3000 {
		t.Fatalf("Wanted aggResp.Months[0].Expense to be 3000, but was %v", aggResp.Months[0])
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

func persistTxn(c context.Context, t *testing.T, txn Txn) {
	if _, err := datastore.Put(c, datastore.NewKey(c, "Txn", "", txn.ID, nil), &txn); err != nil {
		t.Fatal(err)
	}
}
