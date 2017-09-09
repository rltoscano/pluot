package pluot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

// ListTxns lists the transactions in the database filtered by the given parameters.
func listTxns(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	q := datastore.NewQuery("Txn").Order("-PostDate")
	var resp ListTxnsResponse
	keys, err := q.GetAll(c, &resp.Txns)
	if err != nil {
		return nil, err
	}
	for i, k := range keys {
		resp.Txns[i].ID = k.IntID()
	}
	return resp, nil
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
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: err.Error()}
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
			return pihen.RESTErr{Status: http.StatusBadRequest, Message: fmt.Sprintf("`%s` is not an editable field: `userCategory`, `postDate`, `userDisplayName`, `note` ", f)}
		}
	}
	return nil
}
