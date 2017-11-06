package pluot

import (
	"encoding/json"
	"time"
)

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

var (
	// IncomeCategories is a list of all income categories.
	IncomeCategories = []int{
		CategoryPayCheck,
		CategoryBonus,
		CategoryRentals,
		CategoryOtherIncome,
	}
	// ExpenseCategories is a list of all expense categories.
	ExpenseCategories = []int{
		CategoryUncategorized,
		CategoryHomeImprovement,
		CategoryEatingOut,
		CategoryGroceries,
		CategoryLifestyle,
		CategoryHealth,
		CategoryTransportation,
		CategoryResidence,
		CategoryBills,
		CategoryTravel,
		CategoryGifts,
		CategoryOtherExpense,
	}
)

// Upload sources.
const (
	SourceChase      = "chase"
	SourceWellsfargo = "wellsfargo"
)

const (
	// JSONTimeFormat is the string format of JSON UTC time.
	JSONTimeFormat = "Mon, 2 Jan 2006 15:04:05 MST"
)

// Txn represents a financial transaction.
type Txn struct {
	ID                  int64     `datastore:"-" json:"id"`
	PostDate            time.Time `json:"-"`
	Amount              int64     `datastore:"Amount,noindex" json:"amount"`
	OriginalDisplayName string    `json:"originalDisplayName"`
	DisplayName         string    `json:"displayName"`
	UserDisplayName     string    `json:"userDisplayName"`
	Note                string    `json:"note"`
	Category            int       `json:"category"`
	UserCategory        int       `json:"userCategory"`
	UploadID            int64     `json:"uploadId"`
	Splits              []int64   `json:"splits"`
	SplitSourceID       int64     `json:"splitSourceId"`
}

// MarshalJSON marshals a Txn to JSON while converting `PostDate` to a string in JSONTimeFormat.
func (t *Txn) MarshalJSON() ([]byte, error) {
	type Alias Txn
	return json.Marshal(&struct {
		*Alias
		PostDate string `json:"postDate"`
	}{Alias: (*Alias)(t), PostDate: t.PostDate.Format(JSONTimeFormat)})
}

// UnmarshalJSON unmarshals a Txn from JSON while parsing `postDate` to a time.Time.
func (t *Txn) UnmarshalJSON(b []byte) error {
	type Alias Txn
	alias := struct {
		*Alias
		PostDate string `json:"postDate"`
	}{Alias: (*Alias)(t)}
	var err error
	if err = json.Unmarshal(b, &alias); err != nil {
		return err
	}
	if len(alias.PostDate) != 0 {
		t.PostDate, err = time.Parse(JSONTimeFormat, alias.PostDate)
	}
	return err
}

// IsExpenseCategory returns whether the given category is an expense.
func IsExpenseCategory(category int) bool {
	for _, c := range ExpenseCategories {
		if c == category {
			return true
		}
	}
	return false
}

// IsIncomeCategory returns whether the given category is income.
func IsIncomeCategory(category int) bool {
	for _, c := range IncomeCategories {
		if c == category {
			return true
		}
	}
	return false
}
