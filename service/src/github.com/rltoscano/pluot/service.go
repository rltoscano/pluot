package pluot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rltoscano/pihen"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func init() {
	collections := []pihen.RESTCollection{
		{
			URL:           "/svc/txns",
			Methods:       map[string]pihen.RESTMethod{http.MethodGet: listTxns},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/txns/",
			Methods:       map[string]pihen.RESTMethod{http.MethodPatch: patchTxn},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/uploads",
			Methods:       map[string]pihen.RESTMethod{http.MethodPost: createUpload},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/uploads:check",
			Methods:       map[string]pihen.RESTMethod{http.MethodPost: checkUpload},
			AllowedOrigin: "http://localhost:8081",
		},
	}
	pihen.Bind(collections)
	http.HandleFunc("/debug", debugHandler)
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
	t.ID = k.IntID()
	fmt.Fprint(w, "created 1 txn\n")
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(t)
}
