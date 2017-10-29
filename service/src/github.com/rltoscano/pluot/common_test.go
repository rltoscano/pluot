package pluot

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestTxnMarshaling(t *testing.T) {
	postDate, err := time.Parse(JSONTimeFormat, "Sun, 29 Oct 2017 21:28:18 GMT")
	if err != nil {
		t.Fatal(err)
	}
	txn := Txn{
		ID:                  1234,
		Amount:              5678,
		OriginalDisplayName: "originalDisplay",
		DisplayName:         "display",
		UserDisplayName:     "userDisplay",
		Note:                "note",
		Category:            3,
		UserCategory:        4,
		UploadID:            0,
		Splits:              nil,
		SplitSourceID:       0,
		PostDate:            postDate,
	}
	b, err := json.Marshal(&txn)
	if err != nil {
		t.Fatal(err)
	}
	expectedJSON := `{` +
		`"id":1234,` +
		`"amount":5678,` +
		`"originalDisplayName":"originalDisplay",` +
		`"displayName":"display",` +
		`"userDisplayName":"userDisplay",` +
		`"note":"note",` +
		`"category":3,` +
		`"userCategory":4,` +
		`"uploadId":0,` +
		`"splits":null,` +
		`"splitSourceId":0,` +
		`"postDate":"Sun, 29 Oct 2017 21:28:18 GMT"}`
	if string(b) != expectedJSON {
		t.Errorf(`Wanted json to be "%s", but was "%s"`, expectedJSON, string(b))
	}
	unmarshaledTxn := Txn{}
	if err = json.Unmarshal(b, &unmarshaledTxn); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(unmarshaledTxn, txn) {
		t.Errorf("Wanted unmarsaledTxn to be %+v, but was %+v", txn, unmarshaledTxn)
	}
}
