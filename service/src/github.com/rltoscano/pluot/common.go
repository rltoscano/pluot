package pluot

import "time"

// Transaction categories. Don't renumber these.
const (
	CategoryUnknown         = 0
	CategoryUncategorized   = 1
	CategoryHomeImprovement = 2
	CategoryEatingOut       = 3
	CategoryGroceries       = 4
	CategoryLifestyle       = 5
	CategoryHealth          = 6
	CategoryTransportation  = 7
	CategoryResidence       = 8
	CategoryBills           = 9
	CategoryTravel          = 10
	CategoryGifts           = 11
	CategoryOtherExpense    = 12
	CategoryTransfer        = 13
	CategoryPayCheck        = 14
	CategoryBonus           = 15
	CategoryRentals         = 16
	CategoryOtherIncome     = 17
	// This one should be renumbered to always be at the end.
	CategoryEnd = 18
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
