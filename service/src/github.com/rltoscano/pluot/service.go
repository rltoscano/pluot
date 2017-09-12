package pluot

import (
	"net/http"

	"github.com/rltoscano/pihen"
)

func init() {
	collections := []pihen.Collection{
		{
			URL: "/svc/txns",
			Methods: map[string]pihen.Method{
				http.MethodGet:   listTxns,
				http.MethodPatch: patchTxns,
			},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/txns/",
			Methods:       map[string]pihen.Method{http.MethodPatch: patchTxn},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/uploads",
			Methods:       map[string]pihen.Method{http.MethodPost: createUpload},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/uploads:check",
			Methods:       map[string]pihen.Method{http.MethodPost: checkUpload},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL: "/svc/rules",
			Methods: map[string]pihen.Method{
				http.MethodGet:  listRules,
				http.MethodPost: createRule,
			},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/rules:proposal",
			Methods:       map[string]pihen.Method{http.MethodGet: listRuleProposals},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/rules/",
			Methods:       map[string]pihen.Method{http.MethodDelete: deleteRule},
			AllowedOrigin: "http://localhost:8081",
		},
		{
			URL:           "/svc/aggs",
			Methods:       map[string]pihen.Method{http.MethodPost: computeAggregation},
			AllowedOrigin: "http://localhost:8081",
		},
	}
	pihen.Bind(collections)
}
