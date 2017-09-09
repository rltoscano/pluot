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