// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// jsonScalar is an opaque JSON value passed through unchanged. It exists for the
// action ABI (ADR 0029): the SDUI runtime dispatches an Invoke action as
// `mutation($input: JSON) { name(input: $input) }`, so a mutation exposed to
// actions accepts its argument as JSON and interprets it itself.
var jsonScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:         "JSON",
	Description:  "An opaque JSON value passed through unchanged.",
	Serialize:    func(v interface{}) interface{} { return v },
	ParseValue:   func(v interface{}) interface{} { return v },
	ParseLiteral: parseJSONLiteral,
})

// parseJSONLiteral converts an inline GraphQL literal into the Go value a
// variable of the same shape would produce. Clients pass JSON through variables
// (the ParseValue path), so this covers only the rarer inline-literal case.
func parseJSONLiteral(v ast.Value) interface{} {
	switch v := v.(type) {
	case *ast.StringValue:
		return v.Value
	case *ast.BooleanValue:
		return v.Value
	case *ast.IntValue:
		return v.Value
	case *ast.FloatValue:
		return v.Value
	case *ast.ObjectValue:
		m := make(map[string]interface{}, len(v.Fields))
		for _, f := range v.Fields {
			m[f.Name.Value] = parseJSONLiteral(f.Value)
		}
		return m
	case *ast.ListValue:
		l := make([]interface{}, 0, len(v.Values))
		for _, item := range v.Values {
			l = append(l, parseJSONLiteral(item))
		}
		return l
	default:
		return nil
	}
}
