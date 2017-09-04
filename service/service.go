package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

const (
	// Don't renumber these. Next: 6.
	CategoryUncategorized = 0
	CategoryEntertainment = 1
	CategoryEatingOut     = 2
	CategoryGroceries     = 3
	CategoryShopping      = 4
	CategoryHealth        = 5
)

type Txn struct {
	Id                  int64     `datastore:"-" json:"id"`
	PostDate            time.Time `datestore:"post_date" json:"postDate"`
	Amount              int64     `datastore:"amount,noindex" json:"amount"`
	OriginalDisplayName string    `datastore:"original_display_name" json:"originalDisplayName"`
	DisplayName         string    `datastore:"display_name" json:"displayName"`
	UserDisplayName     string    `datastore:"user_display_name" json:"userDisplayName"`
	Note                string    `datastore:"note" json:"note"`
	Category            int       `datastore:"category" json:"category"`
	UserCategory        int       `datastore:"user_category" json:"userCategory"`
}

type TxnList struct {
	Txns []Txn `json:"txns"`
}

func init() {
	http.HandleFunc("/svc/txns", handler)
	http.HandleFunc("/debug", debugHandler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
		return
	}
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("Txn").Order("-PostDate")
	var txns TxnList
	keys, err := q.GetAll(ctx, &txns.Txns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i, k := range keys {
		txns.Txns[i].Id = k.IntID()
	}
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(txns)
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	t := Txn{
		PostDate:            time.Time{},
		Amount:              100,
		OriginalDisplayName: "display name",
		Note:                "note",
		Category:            0,
	}
	k, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "Txn", nil), &t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Id = k.IntID()
	fmt.Fprint(w, "created 1 txn\n")
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(t)
}
