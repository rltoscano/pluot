package pluot

import (
	"net/http"
	"strings"
	"testing"

	"github.com/rltoscano/pihen"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

func TestSplitEmptyBody(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
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
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	r, err := http.NewRequest("method", "url", strings.NewReader(`{ "sourceId": 1, "splits": [] }`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = splitTxn(c, r, nil); err == nil {
		t.Fatal("Expected pihen.Error but got nil")
	}
	if _, ok := err.(pihen.Error); !ok {
		t.Fatal("Expected pihen.Error, but was not")
	}
}

func TestUnknownSource(t *testing.T) {
	c, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	r, err := http.NewRequest(
		"method",
		"url",
		strings.NewReader(`{ "sourceId": 1, "splits": [ { "displayName": "Display Name", "category": 4, "amount": 1000 } ] }`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = splitTxn(c, r, nil); err == nil {
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
	r, err := http.NewRequest(
		"method",
		"url",
		strings.NewReader(`{ "sourceId": 1, "splits": [ { "displayName": "Display Name", "category": 4, "amount": 1000 } ] }`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = splitTxn(c, r, nil); err == nil {
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
	r, err := http.NewRequest(
		"method",
		"url",
		strings.NewReader(`{ "sourceId": 1, "splits": [ { "displayName": "Display Name", "category": 4, "amount": 1000 } ] }`))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := splitTxn(c, r, nil)
	if err != nil {
		t.Fatalf("Expected successful split, but got error %v", err.Error())
	}
	splitResp := resp.(SplitTxnResponse)
	if len(splitResp.Txns) != 1 {
		t.Fatalf("Expected 1 split, but got %v", len(splitResp.Txns))
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
	split1 := Txn{UserDisplayName: "split1", UserCategory: 4, Amount: 400}
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 2, nil), &split1); err != nil {
		t.Fatal(err)
	}
	// Add split 2.
	split2 := Txn{UserDisplayName: "split2", UserCategory: 5, Amount: 600}
	if _, err = datastore.Put(c, datastore.NewKey(c, "Txn", "", 3, nil), &split2); err != nil {
		t.Fatal(err)
	}
	r, err := http.NewRequest(
		"method",
		"url",
		strings.NewReader(`
				{
					"sourceId": 1,
					"splits": [
						{ "displayName": "split1", "category": 4, "amount": 400 },
						{ "displayName": "split3", "category": 7, "amount": 600 }
					]
				}
		`))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := splitTxn(c, r, nil)
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
	if existing.ID != 2 || existing.UserDisplayName != "split1" || existing.UserCategory != 4 || existing.Amount != 400 {
		t.Fatal("Existing transaction unexpectedly modified.")
	}
	if newTxn.UserDisplayName != "split3" || newTxn.UserCategory != 7 || newTxn.Amount != 600 {
		t.Fatal("New transaction has unexpected value.")
	}
}
