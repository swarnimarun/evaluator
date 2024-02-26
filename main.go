package evaluator

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

var (
	SuffixMatch        string = "suffix"
	PrefixMatch        string = "prefix"
	Contains           string = "contains"
	ContainsIgnoreCase string = "contains_ignorecase"
)

type SimpleEvaluator struct {
	ResponseHeaderField string `json:"responseHeaderField,omitempty"`
	RequestHeaderField  string `json:"requestHeaderField,omitempty"`
	Kind                string `json:"kind,omitempty"`
}

func (se SimpleEvaluator) Eval(src, key string) string {
	const TRUE = "true"
	switch se.Kind {
	case SuffixMatch:
		if strings.HasSuffix(src, key) {
			return TRUE
		}
	case PrefixMatch:
		if strings.HasPrefix(src, key) {
			return TRUE
		}
	case Contains:
		if strings.Contains(src, key) {
			return TRUE
		}
	case ContainsIgnoreCase:
		if strings.Contains(
			strings.ToLower(src),
			strings.ToLower(key),
		) {
			return TRUE
		}
	}
	return "false"
}

func (se SimpleEvaluator) IsValid() bool {
	se.Kind = strings.ToLower(se.Kind)
	switch se.Kind {
	case SuffixMatch:
	case PrefixMatch:
	case Contains:
	case ContainsIgnoreCase:
		return true
	}
	return false
}

// add config to select header name
type Config struct {
	SimpleEval       SimpleEvaluator `json:"SimpleEval,omitempty"`
	OutputHeaderName string          `json:"OutputHeaderName,omitempty"`
}

func CreateConfig() *Config {
	return &Config{SimpleEval: SimpleEvaluator{Kind: "", ResponseHeaderField: ""}}
}

type Evaluator struct {
	next   http.Handler
	name   string
	config *Config
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// TODO: also check for invalid header names
	if config.OutputHeaderName == "" {
		return nil, fmt.Errorf("Invalid/Empty OutputHeaderName: '%s'", config.OutputHeaderName)
	}
	if config.SimpleEval.RequestHeaderField == "" {
		return nil, fmt.Errorf("Invalid/Empty SimpleEval.requestHeaderField: '%s'", config.SimpleEval.RequestHeaderField)
	}
	if config.SimpleEval.ResponseHeaderField == "" {
		return nil, fmt.Errorf("Invalid/Empty SimpleEval.responseHeaderField: '%s'", config.SimpleEval.ResponseHeaderField)
	}
	if config.SimpleEval.Kind == "" || !config.SimpleEval.IsValid() {
		return nil, fmt.Errorf("Invalid/Empty SimpleEval.Kind: '%s', valid values are 'suffix', 'prefix', 'contains', 'contains_ignorecase'", config.SimpleEval.Kind)
	}
	return &Evaluator{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

func (a *Evaluator) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	key := req.Header.Get(a.config.SimpleEval.RequestHeaderField)
	a.next.ServeHTTP(resp, req)

	src := resp.Header().Get(a.config.SimpleEval.ResponseHeaderField)
	if src != "" {
		resp.Header().Set(
			a.config.OutputHeaderName,
			a.config.SimpleEval.Eval(src, key),
		)
	}
}
