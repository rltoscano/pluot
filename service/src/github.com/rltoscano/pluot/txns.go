package pluot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rltoscano/pihen"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

// ListTxnsResponse is the response to a list transactions request.
type ListTxnsResponse struct {
	Txns []Txn `json:"txns"`
}

// PatchTxnRequest patches a transaction.
type PatchTxnRequest struct {
	Txn    Txn      `json:"txn"`
	Fields []string `json:"fields"`
}

// PatchTxnsRequest patches multiple transactions.
type PatchTxnsRequest struct {
	IDs    []int64  `json:"ids"`
	Txn    Txn      `json:"txn"`
	Fields []string `json:"fields"`
}

// TxnSplit represents a part of a request to split a transaction. There may be
// multiple splits
type TxnSplit struct {
	DisplayName string `json:"displayName"`
	Category    int    `json:"category"`
	Amount      int64  `json:"amount"`
}

// SplitTxnRequest represents a request to split a source transaction.
type SplitTxnRequest struct {
	SourceID int64      `json:"sourceId"`
	Splits   []TxnSplit `json:"splits"`
}

// SplitTxnResponse contains the newly added split transactions.
type SplitTxnResponse struct {
	Txns []Txn `json:"txns"`
}

// listTxns lists the transactions in the database.
func listTxns(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	txns, err := loadTxns(c, time.Time{}, time.Time{}, CategoryUnknown, false)
	return ListTxnsResponse{txns}, err
}

func patchTxn(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	// Parse ID out of URL.
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return nil, err
	}
	req := PatchTxnRequest{}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	t := new(Txn)
	k := datastore.NewKey(c, "Txn", "", id, nil)
	err = datastore.RunInTransaction(c, func(tc context.Context) error {
		err = datastore.Get(c, k, t)
		if err != nil {
			return err
		}
		if err = applyFields(&req.Txn, t, req.Fields); err != nil {
			return err
		}
		_, err = datastore.Put(c, k, t)
		return err
	}, nil)
	return *t, err
}

func patchTxns(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	// Parse ID out of URL.
	req := PatchTxnsRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	if len(req.IDs) == 0 || len(req.Fields) == 0 {
		return nil, pihen.Error{http.StatusBadRequest, "no input specified"}
	}
	txns := make([]Txn, len(req.IDs))
	keys := make([]*datastore.Key, len(req.IDs))
	for i, id := range req.IDs {
		keys[i] = datastore.NewKey(c, "Txn", "", id, nil)
	}
	err := datastore.RunInTransaction(c, func(tc context.Context) error {
		err := datastore.GetMulti(c, keys, txns)
		if err != nil {
			return err
		}
		for i := range txns {
			if err = applyFields(&req.Txn, &txns[i], req.Fields); err != nil {
				return err
			}
		}
		_, err = datastore.PutMulti(c, keys, txns)
		return err
	}, nil)
	for i, k := range keys {
		txns[i].ID = k.IntID()
	}
	return ListTxnsResponse{txns}, err
}

func splitTxn(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	req := SplitTxnRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	if req.SourceID == 0 || len(req.Splits) == 0 {
		return nil, pihen.Error{http.StatusBadRequest, "no input specified"}
	}

	var source Txn
	sourceKey := datastore.NewKey(c, "Txn", "", req.SourceID, nil)
	splits := make([]Txn, len(req.Splits))
	datastore.RunInTransaction(c, func(tc context.Context) error {
		if err := datastore.Get(tc, sourceKey, &source); err != nil {
			// TODO(robert): Handle if source not found as client error.
			return err
		}
		// TODO(robert): Sanity check splits. E.g. that they add up to the source.
		splitKeys := make([]*datastore.Key, len(req.Splits))
		for i, s := range req.Splits {
			splits[i].PostDate = source.PostDate
			splits[i].Amount = s.Amount
			splits[i].UserDisplayName = s.DisplayName
			splits[i].UserCategory = s.Category
			splits[i].SplitSourceID = source.ID
			splitKeys[i] = datastore.NewIncompleteKey(tc, "Txn", nil)
		}
		splitKeys, err := datastore.PutMulti(tc, splitKeys, &splits)
		if err != nil {
			return err
		}
		source.Splits = make([]int64, len(splitKeys))
		for i, k := range splitKeys {
			source.Splits[i] = k.IntID()
			// Copy key into split for response.
			splits[i].ID = k.IntID()
		}
		_, err = datastore.Put(tc, sourceKey, &source)
		return err
	}, nil)

	return SplitTxnResponse{Txns: splits}, nil
}

func applyFields(source, dest *Txn, fields []string) error {
	for _, f := range fields {
		switch f {
		case "userCategory":
			dest.UserCategory = source.UserCategory
			break
		case "postDate":
			dest.PostDate = source.PostDate
			break
		case "userDisplayName":
			dest.UserDisplayName = source.UserDisplayName
			break
		case "note":
			dest.Note = source.Note
			break
		default:
			return pihen.Error{
				http.StatusBadRequest,
				fmt.Sprintf("`%s` is not an editable field: `userCategory`, `postDate`, `userDisplayName`, `note` ", f),
			}
		}
	}
	return nil
}

// Returns transactions in reverse-chronological order. `end` is exclusive.
func loadTxns(c context.Context, start, end time.Time, cat int, asc bool) ([]Txn, error) {
	q := datastore.NewQuery("Txn")
	if !start.IsZero() {
		q = q.Filter("PostDate >= ", start)
	}
	if !end.IsZero() {
		q = q.Filter("PostDate <", end)
	}
	if asc {
		q = q.Order("PostDate")
	} else {
		q = q.Order("-PostDate")
	}
	txns := []Txn{}
	keys, err := q.GetAll(c, &txns)
	if err != nil {
		return nil, err
	}
	for i, k := range keys {
		txns[i].ID = k.IntID()
	}
	if cat > 0 {
		filtered := make([]Txn, 0, len(txns))
		for _, t := range txns {
			if t.UserCategory > 0 && t.UserCategory != cat {
				continue
			}
			if t.UserCategory == 0 && t.Category != cat {
				continue
			}
			filtered = append(filtered, t)
		}
		txns = filtered
	}
	return txns, nil
}
