package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// specOp is a single OpenAPI operation that carries the Speakeasy group/name
// extensions (the 109 SDK-backed operations).
type specOp struct {
	Group       string
	Name        string
	HTTPMethod  string
	Path        string
	Summary     string
	Description string
}

func parseSpec(path string) ([]specOp, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}

	var ops []specOp
	for p, methods := range doc.Paths {
		for method, raw := range methods {
			if method == "parameters" {
				continue
			}
			var o struct {
				Group       string `json:"x-speakeasy-group"`
				Name        string `json:"x-speakeasy-name-override"`
				Summary     string `json:"summary"`
				Description string `json:"description"`
			}
			if err := json.Unmarshal(raw, &o); err != nil {
				return nil, fmt.Errorf("parse operation %s %s: %w", method, p, err)
			}
			if o.Group == "" || o.Name == "" {
				continue // not an SDK-backed operation; out of scope
			}
			ops = append(ops, specOp{
				Group:       o.Group,
				Name:        o.Name,
				HTTPMethod:  method,
				Path:        p,
				Summary:     o.Summary,
				Description: o.Description,
			})
		}
	}
	return ops, nil
}
