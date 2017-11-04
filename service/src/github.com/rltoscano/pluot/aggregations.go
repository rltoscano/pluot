package pluot

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rltoscano/pihen"

	"golang.org/x/net/context"
	"google.golang.org/appengine/user"
)

// Time windows to aggregate within.
const (
	TimeWindowLast30Days  = 0
	TimeWindowLastMonth   = 1
	TimeWindowLast6Months = 2
)

// ComputeAggregationRequest is a JSON request for the ComputeAggregation
// method.
type ComputeAggregationRequest struct {
	Start time.Time `json:"-"`
	End   time.Time `json:"-"`
}

// MarshalJSON marshals a ComputeAggregationRequest to JSON while converting `Start` and `End` to a
// string in JSONTimeFormat.
func (r *ComputeAggregationRequest) MarshalJSON() ([]byte, error) {
	type Alias ComputeAggregationRequest
	return json.Marshal(&struct {
		*Alias
		Start string `json:"start"`
		End   string `json:"end"`
	}{Alias: (*Alias)(r), Start: r.Start.Format(JSONTimeFormat), End: r.End.Format(JSONTimeFormat)})
}

// UnmarshalJSON unmarshals a ComputeAggregationRequest from JSON while parsing `start` and `end` to
// a time.Time.
func (r *ComputeAggregationRequest) UnmarshalJSON(b []byte) error {
	type Alias ComputeAggregationRequest
	alias := struct {
		*Alias
		Start string `json:"start"`
		End   string `json:"end"`
	}{Alias: (*Alias)(r)}
	var err error
	if err = json.Unmarshal(b, &alias); err != nil {
		return err
	}
	r.Start, err = time.Parse(JSONTimeFormat, alias.Start)
	if err != nil {
		return err
	}
	r.End, err = time.Parse(JSONTimeFormat, alias.End)
	return err
}

// ComputeAggregationResponse contains the total and average aggregations.
type ComputeAggregationResponse struct {
	Totals []int64    `json:"totals"`
	Months []MonthAgg `json:"months"`
}

// MonthAgg contains the aggregation of expenses and income over one month.
type MonthAgg struct {
	Date    time.Time `json:"-"`
	Expense int64     `json:"expense"`
	Income  int64     `json:"income"`
}

// MarshalJSON marshals a MonthAgg to JSON while converting `Date` to a string in JSONTimeFormat.
func (a *MonthAgg) MarshalJSON() ([]byte, error) {
	type Alias MonthAgg
	return json.Marshal(&struct {
		*Alias
		Date string `json:"date"`
	}{Alias: (*Alias)(a), Date: a.Date.Format(JSONTimeFormat)})
}

// UnmarshalJSON unmarshals a MonthAgg from JSON while parsing `date` to a time.Time.
func (a *MonthAgg) UnmarshalJSON(b []byte) error {
	type Alias MonthAgg
	alias := struct {
		*Alias
		Date string `json:"date"`
	}{Alias: (*Alias)(a)}
	var err error
	if err = json.Unmarshal(b, &alias); err != nil {
		return err
	}
	a.Date, err = time.Parse(JSONTimeFormat, alias.Date)
	return err
}

func computeAggregation(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	req := ComputeAggregationRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{Status: http.StatusBadRequest, Message: err.Error()}
	}
	txns, err := loadTxns(c, req.Start, req.End, CategoryUnknown, true)
	if err != nil {
		return nil, err
	}
	resp := ComputeAggregationResponse{
		Totals: make([]int64, CategoryEnd),
		Months: []MonthAgg{},
	}
	currMonth := req.Start
	monthAgg := MonthAgg{Date: currMonth}
	for _, t := range txns {
		if len(t.Splits) > 0 {
			// Don't count split transactions.
			continue
		}
		// Totals.
		cat := t.Category
		if t.UserCategory > 0 {
			cat = t.UserCategory
		}
		resp.Totals[cat] = resp.Totals[cat] + t.Amount
		// Monthly.
		for !currMonth.AddDate(0, 1, 0).After(t.PostDate) {
			resp.Months = append(resp.Months, monthAgg)
			currMonth = currMonth.AddDate(0, 1, 0)
			monthAgg = MonthAgg{Date: currMonth}
		}
		if IsExpenseCategory(cat) {
			monthAgg.Expense = monthAgg.Expense - t.Amount
		}
		if IsIncomeCategory(cat) {
			monthAgg.Income = monthAgg.Income + t.Amount
		}
	}

	resp.Months = append(resp.Months, monthAgg)

	return resp, nil
}
