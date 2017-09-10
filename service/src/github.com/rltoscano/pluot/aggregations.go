package pluot

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rltoscano/pihen"

	"golang.org/x/net/context"
	"google.golang.org/appengine/user"
)

// ComputeAggregationRequest is a JSON request for the ComputeAggregation
// method.
type ComputeAggregationRequest struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Category int    `json:"category"`
}

// ComputeAggregationResponse contains the total and average aggregations.
type ComputeAggregationResponse struct {
	Total   int64 `json:"total"`
	Average int64 `json:"average"`
}

func computeAggregation(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	req := ComputeAggregationRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, pihen.RESTErr{Status: http.StatusBadRequest, Message: err.Error()}
	}
	start, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", req.End)
	if err != nil {
		return nil, err
	}
	txns, err := loadTxns(c, start, end, req.Category)
	if err != nil {
		return nil, err
	}
	resp := ComputeAggregationResponse{}
	for _, t := range txns {
		resp.Total = resp.Total + t.Amount
	}
	if len(txns) > 0 {
		resp.Average = resp.Total / int64(len(txns))
	}
	return resp, nil
}
