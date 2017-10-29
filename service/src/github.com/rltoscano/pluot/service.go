package pluot

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rltoscano/pihen"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

func init() {
	api := pihen.API{
		Collections: []pihen.Collection{
			{
				URL: "/svc/txns",
				Methods: map[string]pihen.Method{
					http.MethodGet:   listTxns,
					http.MethodPatch: patchTxns,
				},
			},
			{
				URL:     "/svc/txns/",
				Methods: map[string]pihen.Method{http.MethodPatch: patchTxn},
			},
			{
				URL:     "/svc/txns:split",
				Methods: map[string]pihen.Method{http.MethodPost: splitTxn},
			},
			{
				URL:     "/svc/uploads",
				Methods: map[string]pihen.Method{http.MethodPost: createUpload},
			},
			{
				URL:     "/svc/uploads:check",
				Methods: map[string]pihen.Method{http.MethodPost: checkUpload},
			},
			{
				URL: "/svc/rules",
				Methods: map[string]pihen.Method{
					http.MethodGet:  listRules,
					http.MethodPost: createRule,
				},
			},
			{
				URL:     "/svc/rules:proposal",
				Methods: map[string]pihen.Method{http.MethodGet: listRuleProposals},
			},
			{
				URL:     "/svc/rules/",
				Methods: map[string]pihen.Method{http.MethodDelete: deleteRule},
			},
			{
				URL:     "/svc/aggs",
				Methods: map[string]pihen.Method{http.MethodPost: computeAggregation},
			},
		},
		AllowedOrigin: "http://localhost:8081",
		Interceptor:   authorizeUser,
	}
	pihen.Bind(api)
}

type config struct {
	// List of comma-separated emails of authorized users.
	AuthorizedUsers string `datastore:"AuthorizedUsers,noindex"`
}

func authorizeUser(c context.Context, r *http.Request, u *user.User) error {
	if u == nil {
		return pihen.Error{Status: http.StatusUnauthorized, Message: "A user could not be determined from the request."}
	}
	var conf config
	k := datastore.NewKey(c, "Config", "singleton", 0, nil)
	err := datastore.Get(c, k, &conf)
	if err == datastore.ErrNoSuchEntity {
		// Create empty config.
		_, err = datastore.Put(c, k, &conf)
	}
	if err != nil {
		return err
	}
	emails := strings.Split(conf.AuthorizedUsers, ",")
	for _, authorized := range emails {
		if authorized == u.Email {
			return nil
		}
	}
	return pihen.Error{Status: http.StatusUnauthorized, Message: fmt.Sprintf("The user %s is unauthorized.", u.Email)}
}
