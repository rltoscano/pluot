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
	splits := make([]Txn, 0, len(req.Splits))
	err := datastore.RunInTransaction(c, func(tc context.Context) error {
		var source Txn
		sourceKey := datastore.NewKey(c, "Txn", "", req.SourceID, nil)
		if err := datastore.Get(tc, sourceKey, &source); err != nil {
			// TODO(robert): Handle if source not found as client error.
			return err
		}
		source.ID = sourceKey.IntID()
		// Check that splits add up to source amount.
		sum := int64(0)
		for _, split := range req.Splits {
			sum = sum + split.Amount
		}
		if sum != source.Amount {
			return pihen.Error{
				http.StatusBadRequest,
				fmt.Sprintf("expected splits to add up to %v, but was %v", source.Amount, sum),
			}
		}
		// Load existing splits (if any).
		oldSplits := make([]Txn, len(source.Splits))
		oldSplitKeys := make([]*datastore.Key, len(source.Splits))
		for i, id := range source.Splits {
			oldSplitKeys[i] = datastore.NewKey(tc, "Txn", "", id, nil)
		}
		if err := datastore.GetMulti(tc, oldSplitKeys, oldSplits); err != nil {
			return fmt.Errorf("could not get existing splits %v for split source %v: %v", source.Splits, source.ID, err)
		}
		for i, k := range oldSplitKeys {
			oldSplits[i].ID = k.IntID()
		}
		// Merge new transactions.
		toDelete := make([]bool, len(oldSplits))
		matches := make([]bool, len(req.Splits))
		for i, oldSplit := range oldSplits {
			match := false
			for j, newSplit := range req.Splits {
				if newSplit.Amount == oldSplit.Amount && newSplit.DisplayName == oldSplit.UserDisplayName && newSplit.Category == oldSplit.UserCategory {
					match = true
					matches[j] = true
					break
				}
			}
			toDelete[i] = !match
		}
		// Delete old unmatched splits.
		deleteKeys := []*datastore.Key{}
		for i, k := range oldSplitKeys {
			if toDelete[i] {
				deleteKeys = append(deleteKeys, k)
			}
		}
		if err := datastore.DeleteMulti(tc, deleteKeys); err != nil {
			return err
		}
		// Add new unmatched splits.
		newSplits := []Txn{}
		newSplitKeys := []*datastore.Key{}
		for i, newSplit := range req.Splits {
			if !matches[i] {
				newSplits = append(newSplits, Txn{
					PostDate:        source.PostDate,
					Amount:          newSplit.Amount,
					UserDisplayName: newSplit.DisplayName,
					UserCategory:    newSplit.Category,
					SplitSourceID:   source.ID,
				})
				newSplitKeys = append(newSplitKeys, datastore.NewIncompleteKey(tc, "Txn", nil))
			}
		}
		newSplitKeys, err := datastore.PutMulti(tc, newSplitKeys, newSplits)
		if err != nil {
			return err
		}
		for i, k := range newSplitKeys {
			newSplits[i].ID = k.IntID()
		}
		// Update source.
		for i, oldSplit := range oldSplits {
			if !toDelete[i] {
				splits = append(splits, oldSplit)
			}
		}
		for _, newSplit := range newSplits {
			splits = append(splits, newSplit)
		}
		source.Splits = make([]int64, len(splits))
		for i, split := range splits {
			source.Splits[i] = split.ID
		}
		_, err = datastore.Put(tc, sourceKey, &source)
		return err
	}, &datastore.TransactionOptions{XG: true})
	if err != nil {
		return nil, err
	}
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
