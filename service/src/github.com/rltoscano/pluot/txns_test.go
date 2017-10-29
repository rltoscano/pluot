package pluot

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rltoscano/pihen"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestCreateTxn(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	requestBody := `{
		"txn":{
			"postDate": "Sun, 29 Oct 2017 21:28:18 GMT",
			"amount": 123,
			"userDisplayName": "display",
			"userCategory": 5,
			"note": "my note"
		}
	}`
	resp, err := createTxn(c, newRequest(t, requestBody), nil)
	if err != nil {
		t.Fatal(err)
	}
	created := resp.(Txn)
	expectedPostDate, err := time.Parse(JSONTimeFormat, "Sun, 29 Oct 2017 21:28:18 GMT")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 {
		t.Error("Wanted non-zero created.ID, but was 0")
	}
	if !created.PostDate.Equal(expectedPostDate) {
		t.Errorf("Wanted created.PostDate to be %s, but was %s", expectedPostDate, created.PostDate)
	}
	if created.Amount != 123 {
		t.Errorf("Wanted created.Amount 123, but was %v", created.Amount)
	}
	if created.UserDisplayName != "display" {
		t.Errorf("Wanted created.UserDisplayName to be display, but was %s", created.UserDisplayName)
	}
	if created.UserCategory != 5 {
		t.Errorf("Wanted created.UserCategory to be 5, but was %v", created.UserCategory)
	}
	if created.Note != "my note" {
		t.Errorf("Wanted created.Note to be my note, but was %s", created.Note)
	}

	// Allow time for database write.
	d, _ := time.ParseDuration("500ms")
	time.Sleep(d)

	stored := Txn{}
	if err = datastore.Get(c, datastore.NewKey(c, "Txn", "", created.ID, nil), &stored); err != nil {
		t.Fatal(err)
	}
	if !stored.PostDate.Equal(created.PostDate) {
		t.Errorf("Wanted stored.PostDate to be %s, but was %s", expectedPostDate, stored.PostDate)
	}
	if stored.Amount != created.Amount {
		t.Errorf("Wanted stored.Amount to be %v, but was %v", created.Amount, stored.Amount)
	}
	if stored.UserDisplayName != created.UserDisplayName {
		t.Errorf("Wanted stored.UserDisplayName to be %s, but was %s", created.UserDisplayName, stored.UserDisplayName)
	}
	if stored.UserCategory != created.UserCategory {
		t.Errorf("Wanted stored.UserCategory to be %v, but was %v", created.UserCategory, stored.UserCategory)
	}
	if stored.Note != created.Note {
		t.Errorf("Wanted stored.Note to be %s, but was %s", created.Note, stored.Note)
	}
}

func TestCreateTxnInvalidRequests(t *testing.T) {
	bodies := []string{
		`--`,
		`{}`,
		`{"txn":{"postDate":"Sun, 29 Oct 2017 21:28:18 GMT", "userDisplayName":"d", "userCategory":1}}`,
		`{"txn":{"amount":0, "postDate":"Sun, 29 Oct 2017 21:28:18 GMT", "userDisplayName":"d", "userCategory":1}}`,
		`{"txn":{"amount":123, "userDisplayName": "d", "userCategory":1}}`,
		`{"txn":{"amount":123, "postDate":"Sun, 29 Oct 2017 21:28:18 GMT", "userCategory":1}}`,
		`{"txn":{"amount":123, "postDate":"Sun, 29 Oct 2017 21:28:18 GMT", "userDisplayName":"d"}}`,
	}
	c := context.Background()
	for _, body := range bodies {
		_, err := createTxn(c, newRequest(t, body), nil)
		if err == nil {
			t.Fatalf(`Wanted error from createTxn("%s"), but there was no error`, body)
		}
		pErr, ok := err.(pihen.Error)
		if !ok {
			t.Fatalf("Wanted pihen.Error, but got %v", reflect.TypeOf(err))
		}
		if pErr.Status != http.StatusBadRequest {
			t.Fatalf("Wanted status 400, but got %v", pErr.Status)
		}
	}
}

func TestSplitEmptyBody(t *testing.T) {
	c := context.Background()
	r, err := http.NewRequest("method", "url", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = splitTxn(c, r, nil); err == nil {
		t.Fatal("Expected pihen.Error but got nil")
	}
	pErr, ok := err.(pihen.Error)
	if !ok {
		t.Fatal("Expected pihen.Error, but was not")
	}
	if pErr.Status != http.StatusBadRequest {
		t.Fatalf("Wanted BadRequest status, but was %v", pErr.Status)
	}
}

func TestNoSplits(t *testing.T) {
	c := context.Background()
	_, err := splitTxn(c, newRequest(t, `{ "sourceId": 1, "splits": [] }`), nil)
	if err == nil {
		t.Fatal("Expected pihen.Error but got nil")
	}
	if _, ok := err.(pihen.Error); !ok {
		t.Fatalf("Expected pihen.Error, but was %v", reflect.TypeOf(err))
	}
}

func TestUnknownSource(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	body := `{ "sourceId": 1, "splits": [ { "displayName": "Display Name", "category": 4, "amount": 1000 } ] }`
	if _, err = splitTxn(c, newRequest(t, body), nil); err == nil {
		t.Fatal("Expected error but got nil")
	}
}

func TestSplitSumInvalid(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	// Add source.
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 1, nil), &Txn{Amount: 999}); err != nil {
		t.Fatal(err)
	}
	body := `{ "sourceId": 1, "splits": [ { "displayName": "Display Name", "category": 4, "amount": 1000 } ] }`
	if _, err = splitTxn(c, newRequest(t, body), nil); err == nil {
		t.Fatalf("Expected error, but was successful")
	}
}

func TestSingleSplit(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	// Add source.
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 1, nil), &Txn{Amount: 1000}); err != nil {
		t.Fatal(err)
	}
	body := `{ "sourceId": 1, "splits": [ { "displayName": "Display Name", "category": 4, "amount": 1000 } ] }`
	resp, err := splitTxn(c, newRequest(t, body), nil)
	if err != nil {
		t.Fatalf("Expected successful split, but got error %v", err.Error())
	}
	splitResp := resp.(SplitTxnResponse)
	if len(splitResp.Txns) != 1 {
		t.Fatalf("Expected 1 split, but got %v", len(splitResp.Txns))
	}
	if splitResp.Txns[0].SplitSourceID != 1 {
		t.Fatalf("Expected split to have source ID 1, but was %v", splitResp.Txns[0].SplitSourceID)
	}
}

func TestResplit(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	// Add source.
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 1, nil), &Txn{Amount: 1000, Splits: []int64{2, 3}}); err != nil {
		t.Fatal(err)
	}
	// Add split 1.
	split1 := Txn{UserDisplayName: "split1", UserCategory: 4, Amount: 400, SplitSourceID: 1}
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 2, nil), &split1); err != nil {
		t.Fatal(err)
	}
	// Add split 2.
	split2 := Txn{UserDisplayName: "split2", UserCategory: 5, Amount: 600, SplitSourceID: 1}
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 3, nil), &split2); err != nil {
		t.Fatal(err)
	}
	body := `
		{
			"sourceId": 1,
			"splits": [
				{ "displayName": "split1", "category": 4, "amount": 400 },
				{ "displayName": "split3", "category": 7, "amount": 600 }
			]
		}`
	resp, err := splitTxn(c, newRequest(t, body), nil)
	if err != nil {
		t.Fatalf("Expected successful split, but got error %v", err.Error())
	}
	splitResp := resp.(SplitTxnResponse)
	if len(splitResp.Txns) != 2 {
		t.Fatalf("Expected 1 split, but got %v", len(splitResp.Txns))
	}
	// Verify existing txn is still there.
	var existing Txn
	var newTxn Txn
	if splitResp.Txns[0].ID == 2 {
		existing = splitResp.Txns[0]
		newTxn = splitResp.Txns[1]
	} else {
		existing = splitResp.Txns[1]
		newTxn = splitResp.Txns[0]
	}
	if existing.ID != 2 || existing.UserDisplayName != "split1" || existing.UserCategory != 4 || existing.Amount != 400 || existing.SplitSourceID != 1 {
		t.Fatal("Existing transaction unexpectedly modified.")
	}
	if newTxn.UserDisplayName != "split3" || newTxn.UserCategory != 7 || newTxn.Amount != 600 || newTxn.SplitSourceID != 1 {
		t.Fatal("New transaction has unexpected value.")
	}
}

func newRequest(t *testing.T, body string) *http.Request {
	r, err := http.NewRequest("method", "url", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
