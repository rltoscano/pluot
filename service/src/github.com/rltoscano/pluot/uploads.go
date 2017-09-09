package pluot

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rltoscano/pihen"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// UploadEvent represents an transaction upload event.
type UploadEvent struct {
	ID        int64     `datastore:"-" json:"id"`
	Source    string    `datastore:"source" json:"source"`
	User      string    `datastore:"user" json:"user"`
	Count     int       `datastore:"count" json:"count"`
	EventTime time.Time `datastore:"event_time" json:"eventTime"`
	Start     time.Time `datastore:"start" json:"start"`
	End       time.Time `datastore:"end" json:"end"`
}

// UploadDuplicate represents a duplication of a transaction observed in an
// upload.
type UploadDuplicate struct {
	DupID       int64     `json:"dupId"`
	UploadIdx   int       `json:"uploadIdx"`
	PostDate    time.Time `json:"postDate"`
	DisplayName string    `json:"displayName"`
	Amount      int64     `json:"amount"`
}

// CheckUploadResponse represents the response to a CheckUpload request.
type CheckUploadResponse struct {
	Duplicates []UploadDuplicate `json:"duplicates"`
}

func checkUpload(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	// Parse and validate input.
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	if len(r.Form["source"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `source` parameter"}
	}
	if len(r.Form["csv"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `csv` parameter"}
	}
	if len(r.Form["start"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `start` parameter"}
	}
	if len(r.Form["end"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `end` parameter"}
	}
	uploadedTxns, err := parseTxns(r.Form["csv"][0], r.Form["source"][0])
	if err != nil {
		return nil, err
	}
	start, err := time.Parse("2006-01-02", r.Form["start"][0])
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", r.Form["end"][0])
	if err != nil {
		return nil, err
	}

	// Load existing transactions.
	q := datastore.NewQuery("Txn").
		Filter("PostDate <", end).
		Filter("PostDate >", start)
	existingTxns := []Txn{}
	keys, err := q.GetAll(c, &existingTxns)
	if err != nil {
		return nil, err
	}
	for i, k := range keys {
		existingTxns[i].ID = k.IntID()
	}
	log.Debugf(c, "%d transactions loaded", len(existingTxns))

	// Compare transactions.
	duplicates := make([]UploadDuplicate, 0, len(existingTxns))
	for i, u := range uploadedTxns {
		for _, e := range existingTxns {
			if u.Amount == e.Amount && u.PostDate == e.PostDate && u.OriginalDisplayName == e.OriginalDisplayName {
				duplicates = append(duplicates, UploadDuplicate{
					DupID:       e.ID,
					UploadIdx:   i,
					PostDate:    u.PostDate,
					DisplayName: u.OriginalDisplayName,
					Amount:      u.Amount,
				})
				break
			}
		}
	}

	return CheckUploadResponse{duplicates}, nil
}

func createUpload(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	if len(r.Form["source"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `source` parameter"}
	}
	if len(r.Form["csv"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `csv` parameter"}
	}
	if len(r.Form["start"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `start` parameter"}
	}
	if len(r.Form["end"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `end` parameter"}
	}
	if len(r.Form["ignore"]) == 0 {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: "missing `ignore` parameter"}
	}
	sourceName := r.Form["source"][0]
	uploadedTxns, err := parseTxns(r.Form["csv"][0], sourceName)
	if err != nil {
		return nil, err
	}
	start, err := time.Parse("2006-01-02", r.Form["start"][0])
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", r.Form["end"][0])
	if err != nil {
		return nil, err
	}

	// TODO(robert): Filter out duplicates.
	// ignore := r.Form["ignore"][0]
	filteredTxns := uploadedTxns

	// Record upload event.
	// TODO(robert): Make this part of a transaction with above mutation.
	createUploadResp := UploadEvent{
		Source:    sourceName,
		User:      "testing",
		Count:     len(filteredTxns),
		EventTime: time.Now(),
		Start:     start,
		End:       end,
	}
	k, err := datastore.Put(c, datastore.NewIncompleteKey(c, "UploadEvent", nil), &createUploadResp)
	if err != nil {
		return nil, err
	}
	createUploadResp.ID = k.IntID()

	for i := range filteredTxns {
		filteredTxns[i].UploadID = createUploadResp.ID
	}

	// Record transactions.
	keys := make([]*datastore.Key, len(filteredTxns))
	for i := range keys {
		keys[i] = datastore.NewIncompleteKey(c, "Txn", nil)
	}
	if _, err = datastore.PutMulti(c, keys, filteredTxns); err != nil {
		return nil, err
	}

	return createUploadResp, nil
}

func parseTxns(csvStr string, source string) ([]Txn, error) {
	csvRows, err := csv.NewReader(strings.NewReader(csvStr)).ReadAll()
	if err != nil {
		return nil, err
	}
	switch source {
	case SourceChase:
		return parseChase(csvRows)
	case SourceWellsfargo:
		return parseWellsfargo(csvRows)
	default:
		return nil, fmt.Errorf("unexpected source %v", source)
	}
}

// Type, TransDate, PostDate, Description, Amount
// First row is headers.
func parseChase(csvRows [][]string) ([]Txn, error) {
	txns := make([]Txn, 0, len(csvRows)-1)
	for _, row := range csvRows[1:] {
		postDate, err := time.Parse("01/02/2006", row[2])
		if err != nil {
			return nil, err
		}
		f, err := strconv.ParseFloat(row[4], 32)
		if err != nil {
			return nil, err
		}
		txns = append(txns, Txn{
			PostDate:            postDate,
			OriginalDisplayName: row[3],
			Amount:              int64(f * 100),
		})
	}
	return txns, nil
}

func parseWellsfargo(csvRows [][]string) ([]Txn, error) {
	return nil, nil
}
