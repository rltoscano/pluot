package pluot

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

// ListTxnsResponse is the response to a list transactions request.
type ListTxnsResponse struct {
	Txns []Txn `json:"txns"`
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
