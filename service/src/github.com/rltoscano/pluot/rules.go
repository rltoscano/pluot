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

// Rule applies a regular expression to a transaction's display name and if
// it matches, changes the transaction's category.
type Rule struct {
	ID       int64     `json:"id"`
	Regexp   string    `json:"regexp"`
	Category int       `json:"category"`
	Created  time.Time `json:"created"`
}

// Applies returns whether the rule applies to the transaction.
func (r Rule) Applies(c context.Context, t Txn) bool {
	matched, err := regexp.MatchString(r.Regexp, t.OriginalDisplayName)
	if err != nil {
		log.Errorf(c, "error applying rule regexp `%s` to string `%s`", r.Regexp, t.OriginalDisplayName)
		return false
	}
	return matched
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
	if rule.Regexp == "" {
		return nil, pihen.Error{http.StatusBadRequest, "missing `regexp` parameter"}
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
