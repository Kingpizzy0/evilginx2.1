package core

import (
	"regexp"
	"strings"
	"time"
)

type UrlRewriteQuery struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type UrlRewrite struct {
	Path        string            `yaml:"path"`
	QueryParams []UrlRewriteQuery `yaml:"query_params"`
}

type UrlRewriteTrigger struct {
	Domains []string `yaml:"domains"`
	Paths   []string `yaml:"paths"`
}

type UrlRewriteRule struct {
	Trigger UrlRewriteTrigger `yaml:"trigger"`
	Rewrite UrlRewrite        `yaml:"rewrite"`
}

type UrlRewriter struct {
	rules []*UrlRewriteRule
}

func NewUrlRewriter() *UrlRewriter {
	return &UrlRewriter{
		rules: make([]*UrlRewriteRule, 0),
	}
}

// AddRule adds a URL rewriting rule
func (u *UrlRewriter) AddRule(rule *UrlRewriteRule) {
	u.rules = append(u.rules, rule)
}

// RewriteUrl checks if a URL should be rewritten and returns the new URL
func (u *UrlRewriter) RewriteUrl(domain string, path string, sessionId string, originalPath string) (string, bool) {
	for _, rule := range u.rules {
		// Check if domain matches
		domainMatch := false
		for _, d := range rule.Trigger.Domains {
			if d == domain {
				domainMatch = true
				break
			}
		}
		
		if !domainMatch {
			continue
		}
		
		// Check if path matches any regex pattern
		pathMatch := false
		for _, p := range rule.Trigger.Paths {
			re, err := regexp.Compile(p)
			if err != nil {
				continue
			}
			if re.MatchString(path) {
				pathMatch = true
				break
			}
		}
		
		if !pathMatch {
			continue
		}
		
		// Build rewritten URL
		newPath := rule.Rewrite.Path
		
		// Build query parameters
		queryParams := make([]string, 0)
		for _, qp := range rule.Rewrite.QueryParams {
			key := qp.Key
			value := qp.Value
			
			// Replace placeholders
			value = strings.ReplaceAll(value, "{session_id}", sessionId)
			value = strings.ReplaceAll(value, "{original_path}", originalPath)
			value = strings.ReplaceAll(value, "{timestamp}", string(time.Now().Unix()))
			value = strings.ReplaceAll(value, "{id}", sessionId) // Legacy support
			
			queryParams = append(queryParams, key+"="+value)
		}
		
		// Construct final URL
		if len(queryParams) > 0 {
			newPath += "?" + strings.Join(queryParams, "&")
		}
		
		return newPath, true
	}
	
	return path, false
}

// GetOriginalPath extracts the original path from rewritten URL query parameters
func (u *UrlRewriter) GetOriginalPath(queryParams map[string]string) (string, bool) {
	if originalPath, ok := queryParams["redirect"]; ok {
		return originalPath, true
	}
	return "", false
}
