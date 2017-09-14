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
	TimeWindow int `json:"timeWindow"`
}

// ComputeAggregationResponse contains the total and average aggregations.
type ComputeAggregationResponse struct {
	Totals []int64 `json:"totals"`
}

func computeAggregation(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	req := ComputeAggregationRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	var start, end time.Time
	switch req.TimeWindow {
	case TimeWindowLast30Days:
		end = time.Now()
		start = end.Add(-time.Hour * 24 * 30)
		break
	case TimeWindowLastMonth:
		now := time.Now()
		end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		start = end.AddDate(0, -1, 0)
		break
	case TimeWindowLast6Months:
		now := time.Now()
		end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		start = end.AddDate(0, -6, 0)
		break
	}
	txns, err := loadTxns(c, start, end, CategoryUnknown)
	if err != nil {
		return nil, err
	}
	resp := ComputeAggregationResponse{
		Totals: make([]int64, CategoryEnd),
	}
	for _, t := range txns {
		cat := t.Category
		if t.UserCategory > 0 {
			cat = t.UserCategory
		}
		resp.Totals[cat] = resp.Totals[cat] + t.Amount
	}
	return resp, nil
}
