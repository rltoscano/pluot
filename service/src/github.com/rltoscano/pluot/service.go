package pluot

import (
	"net/http"

	"github.com/rltoscano/pihen"
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
		AllowedOrigin:      "http://localhost:8081",
		UserEmailWhitelist: []string{"test@example.com"},
	}
	pihen.Bind(api)
}
