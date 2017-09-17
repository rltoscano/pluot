package pluot

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rltoscano/pihen"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

// Rule tries to match against transactions in order to transform their display
// name and/or category. Both whole string and regexp matching are supported.
type Rule struct {
	ID          int64     `json:"id"`
	Pattern     string    `json:"pattern"`
	Regexp      bool      `json:"regexp"`
	DisplayName string    `json:"displayName"`
	Category    int       `json:"category"`
	Created     time.Time `json:"created"`
}

// ListRuleProposalsResponse contains the list of rule proposals.
type ListRuleProposalsResponse struct {
	Rules []Rule `json:"rules"`
}

// Applies returns whether the rule applies to the transaction.
func (r Rule) Applies(c context.Context, t Txn) bool {
	if r.Regexp {
		matched, err := regexp.MatchString(r.Pattern, t.OriginalDisplayName)
		if err != nil {
			log.Errorf(c, "error applying rule regexp `%s` to string `%s`", r.Pattern, t.OriginalDisplayName)
			return false
		}
		return matched
	}
	return r.Pattern == t.OriginalDisplayName
}

// ListRulesResponse is returned when listing rules.
type ListRulesResponse struct {
	Rules []Rule `json:"rules"`
}

func createRule(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	var rule Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	if rule.Pattern == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `pattern` parameter"}
	}
	if rule.Category == 0 {
		return nil, pihen.Error{http.StatusBadRequest, "invalid `category`"}
	}
	rule.Created = time.Now()
	k, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Rule", nil), &rule)
	if err != nil {
		return nil, err
	}
	rule.ID = k.IntID()
	return rule, nil
}

func deleteRule(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	if err != nil {
		return nil, pihen.Error{http.StatusBadRequest, err.Error()}
	}
	return struct{}{}, datastore.Delete(c, datastore.NewKey(c, "Rule", "", id, nil))
}

func listRules(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	rules, err := loadRules(c)
	return ListRulesResponse{rules}, err
}

func listRuleProposals(c context.Context, r *http.Request, u *user.User) (interface{}, error) {
	end := time.Now()
	start := end.Add(-time.Hour * 24 * 30 * 6)
	txns, err := loadTxns(c, start, end, CategoryUnknown, false)
	if err != nil {
		return nil, err
	}
	proposals := []Rule{}
	for _, t := range txns {
		if t.UserDisplayName == "" && t.UserCategory == CategoryUnknown {
			continue
		}
		r := Rule{Pattern: t.OriginalDisplayName, Regexp: false}
		if t.UserDisplayName != "" {
			r.DisplayName = t.UserDisplayName
		}
		if t.UserCategory != CategoryUnknown {
			r.Category = t.UserCategory
		}
		proposals = append(proposals, r)
		if len(proposals) >= 5 {
			break
		}
	}
	return ListRuleProposalsResponse{proposals}, nil
}

// loadRules loads rules from storage.
func loadRules(c context.Context) ([]Rule, error) {
	q := datastore.NewQuery("Rule")
	rules := []Rule{}
	keys, err := q.GetAll(c, &rules)
	if err != nil {
		return nil, err
	}
	for i, k := range keys {
		rules[i].ID = k.IntID()
	}
	return rules, nil
}
