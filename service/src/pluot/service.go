package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// Transaction categories. Don't renumber these. Next: 6.
const (
	CategoryUncategorized = 0
	CategoryEntertainment = 1
	CategoryEatingOut     = 2
	CategoryGroceries     = 3
	CategoryShopping      = 4
	CategoryHealth        = 5
)

// Upload sources.
const (
	SourceChase      = "chase"
	SourceWellsfargo = "wellsfargo"
)

// Txn represents a financial transaction.
type Txn struct {
	ID                  int64     `datastore:"-" json:"id"`
	PostDate            time.Time `datestore:"post_date" json:"postDate"`
	Amount              int64     `datastore:"amount,noindex" json:"amount"`
	OriginalDisplayName string    `datastore:"original_display_name" json:"originalDisplayName"`
	DisplayName         string    `datastore:"display_name" json:"displayName"`
	UserDisplayName     string    `datastore:"user_display_name" json:"userDisplayName"`
	Note                string    `datastore:"note" json:"note"`
	Category            int       `datastore:"category" json:"category"`
	UserCategory        int       `datastore:"user_category" json:"userCategory"`
	UploadID            int64     `datastore:"upload_id" json:"uploadId"`
}

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

// ListTxnsResponse is the response to a list transactions request.
type ListTxnsResponse struct {
	Txns []Txn `json:"txns"`
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

func init() {
	http.HandleFunc("/svc/txns", listTxns)
	http.HandleFunc("/svc/uploads:check", checkUpload)
	http.HandleFunc("/svc/uploads", createUpload)
	http.HandleFunc("/debug", debugHandler)
}

func listTxns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
		return
	}
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("Txn").Order("-PostDate")
	var txns ListTxnsResponse
	keys, err := q.GetAll(ctx, &txns.Txns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i, k := range keys {
		txns.Txns[i].ID = k.IntID()
	}
	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(txns)
}

func checkUpload(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sourceName := r.Form["source"][0]
	uploadedTxns, err := parseTxns(r.Form["csv"][0], sourceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// Load existing transactions.
	start, err := time.Parse("2006-01-02", r.Form["start"][0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	end, err := time.Parse("2006-01-02", r.Form["end"][0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	q := datastore.NewQuery("Txn").
		Filter("PostDate <", end).
		Filter("PostDate >", start).
		Order("-PostDate")
	existingTxns := make([]Txn, 0)
	keys, err := q.GetAll(ctx, &existingTxns)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i, k := range keys {
		existingTxns[i].ID = k.IntID()
	}

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

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(CheckUploadResponse{duplicates})
}

func createUpload(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	sourceName := r.Form["source"][0]
	uploadedTxns, err := parseTxns(r.Form["csv"][0], sourceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	start, err := time.Parse("2006-01-02", r.Form["start"][0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	end, err := time.Parse("2006-01-02", r.Form["end"][0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
	k, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "UploadEvent", nil), &createUploadResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	createUploadResp.ID = k.IntID()

	for i := range filteredTxns {
		filteredTxns[i].UploadID = createUploadResp.ID
	}

	// Record transactions.
	keys := make([]*datastore.Key, len(filteredTxns))
	for i := range keys {
		keys[i] = datastore.NewIncompleteKey(ctx, "Txn", nil)
	}
	if _, err = datastore.PutMulti(ctx, keys, filteredTxns); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(createUploadResp)
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
