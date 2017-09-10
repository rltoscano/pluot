package pluot

import "time"

// Transaction categories. Don't renumber these. Next: 6.
const (
	CategoryUnknown       = 0
	CategoryUncategorized = 1
	CategoryEntertainment = 2
	CategoryEatingOut     = 3
	CategoryGroceries     = 4
	CategoryShopping      = 5
	CategoryHealth        = 6
)

// Upload sources.
const (
	SourceChase      = "chase"
	SourceWellsfargo = "wellsfargo"
)

// Txn represents a financial transaction.
type Txn struct {
	ID                  int64     `datastore:"-" json:"id"`
	PostDate            time.Time `json:"postDate"`
	Amount              int64     `datastore:"Amount,noindex" json:"amount"`
	OriginalDisplayName string    `json:"originalDisplayName"`
	DisplayName         string    `json:"displayName"`
	UserDisplayName     string    `json:"userDisplayName"`
	Note                string    `json:"note"`
	Category            int       `json:"category"`
	UserCategory        int       `json:"userCategory"`
	UploadID            int64     `json:"uploadId"`
}
