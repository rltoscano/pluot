package pluot

import (
	"encoding/json"
	"fmt"
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
	TimeWindow int    `json:"timeWindow"`
	Start      string `json:"start"`
	End        string `json:"end"`
}

// ComputeAggregationResponse contains the total and average aggregations.
type ComputeAggregationResponse struct {
	Totals []int64    `json:"totals"`
	Months []MonthAgg `json:"months"`
}

// MonthAgg contains the aggregation of expenses and income over one month.
type MonthAgg struct {
	Month   string `json:"month"`
	Expense int64  `json:"expense"`
	Income  int64  `json:"income"`
}

func computeAggregation(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	req := ComputeAggregationRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{Status: http.StatusBadRequest, Message: err.Error()}
	}
	var start, end time.Time
	if req.Start != "" && req.End != "" {
		var err error
		start, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", req.Start)
		if err != nil {
			return nil, pihen.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid `start` value: %v", err)}
		}
		end, err = time.Parse("Mon, 02 Jan 2006 15:04:05 MST", req.End)
		if err != nil {
			return nil, pihen.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("Invalid `end` value: %v", err)}
		}
	} else {
		pst, _ := time.LoadLocation("America/Los_Angeles")
		switch req.TimeWindow {
		case TimeWindowLast30Days:
			end = time.Now()
			start = end.Add(-time.Hour * 24 * 30)
			break
		case TimeWindowLastMonth:
			now := time.Now()
			end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, pst)
			start = end.AddDate(0, -1, 0)
			break
		case TimeWindowLast6Months:
			now := time.Now()
			end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, pst)
			start = end.AddDate(0, -6, 0)
			break
		}
	}
	txns, err := loadTxns(c, start, end, CategoryUnknown, true)
	if err != nil {
		return nil, err
	}
	resp := ComputeAggregationResponse{
		Totals: make([]int64, CategoryEnd),
		Months: []MonthAgg{},
	}
	currMonth := start
	monthAgg := MonthAgg{Month: monthStr(currMonth)}
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
			monthAgg = MonthAgg{Month: monthStr(currMonth)}
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

func monthStr(t time.Time) string {
	return t.Format("Jan 2006")
}
