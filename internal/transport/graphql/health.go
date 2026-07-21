// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package graphql

import "github.com/graphql-go/graphql"

// componentHealthType is a schema-shape placeholder for domain.HealthStatus.
// A per-adapter contracts.HealthProbe already exists (e.g. Postgres's), but
// there is no cross-component aggregation and no application service exposing
// any of it yet — the diagnostics model is deferred.
var componentHealthType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ComponentHealth",
	Fields: graphql.Fields{
		"component": &graphql.Field{Type: graphql.String},
		"state":     &graphql.Field{Type: graphql.String},
		"detail":    &graphql.Field{Type: graphql.String},
	},
})

// healthGap is the flagged reason the Health field below fails — see
// componentHealthType.
const healthGap = "component health aggregation is not implemented yet (contracts.HealthProbe exists per-adapter, e.g. Postgres's, but no cross-component application service exists — that is the Diagnostics and health slice's job)"

// componentHealthField stubs the diagnostics model's "component health and
// degraded component detail" query — see healthGap.
func componentHealthField() *graphql.Field {
	return notImplementedField(graphql.NewList(componentHealthType), graphql.FieldConfigArgument{
		"callerSessionId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
	}, healthGap)
}
