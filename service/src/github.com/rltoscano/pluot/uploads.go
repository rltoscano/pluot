package pluot

import (
	"encoding/csv"
	"encoding/json"
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
	Source    string    `json:"source"`
	User      string    `json:"user"`
	Count     int       `json:"count"`
	EventTime time.Time `json:"eventTime"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
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

// UploadRequest represents both a CheckUpload and CreateUpload request.
type UploadRequest struct {
	Source string `json:"source"`
	CSV    string `json:"csv"`
	Start  string `json:"start"`
	End    string `json:"end"`
	Ignore []int  `json:"ignore"`
}

// CheckUploadResponse represents the response to a CheckUpload request.
type CheckUploadResponse struct {
	Duplicates []UploadDuplicate `json:"duplicates"`
}

func checkUpload(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	// Parse and validate input.
	req := UploadRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	if req.Source == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `source` parameter"}
	}
	if req.CSV == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `csv` parameter"}
	}
	if req.Start == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `start` parameter"}
	}
	if req.End == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `end` parameter"}
	}
	uploadedTxns, err := parseTxns(req.CSV, req.Source)
	if err != nil {
		return nil, err
	}
	start, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", req.End)
	if err != nil {
		return nil, err
	}

	// Load existing transactions.
	existingTxns, err := loadTxns(c, start, end, CategoryUnknown)
	if err != nil {
		return nil, err
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
	req := UploadRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	if req.Source == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `source` parameter"}
	}
	if req.CSV == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `csv` parameter"}
	}
	if req.Start == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `start` parameter"}
	}
	if req.End == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `end` parameter"}
	}
	uploadedTxns, err := parseTxns(req.CSV, req.Source)
	if err != nil {
		return nil, err
	}
	start, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", req.End)
	if err != nil {
		return nil, err
	}

	// Filter out duplicates.
	filteredTxns := make([]Txn, 0, len(uploadedTxns))
	for i, t := range uploadedTxns {
		skip := false
		for _, ignore := range req.Ignore {
			if i == ignore {
				skip = true
				break
			}
		}
		if !skip {
			filteredTxns = append(filteredTxns, t)
		}
	}

	// Apply rules.
	rules, err := loadRules(c)
	if err != nil {
		return nil, err
	}
	for i, t := range filteredTxns {
		for _, r := range rules {
			if r.Applies(c, t) {
				if r.DisplayName != "" {
					filteredTxns[i].DisplayName = r.DisplayName
				}
				if r.Category != CategoryUnknown {
					filteredTxns[i].Category = r.Category
				}
				break
			}
		}
	}

	// Record upload event.
	// TODO(robert): Make this part of a transaction with above mutation.
	// TODO(robert): Add user support.
	createUploadResp := UploadEvent{
		Source:    req.Source,
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
			Category:            CategoryUncategorized,
		})
	}
	return txns, nil
}

func parseWellsfargo(csvRows [][]string) ([]Txn, error) {
	return nil, nil
}
