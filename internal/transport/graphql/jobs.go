// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package graphql

import "github.com/graphql-go/graphql"

// jobType is a schema-shape placeholder matching the columns migration
// 0006_jobs.sql created (id, kind, status, ...). No domain type, contract, or
// application service exists on top of that table yet, so every Jobs resolver
// below is a flagged stub, not a real projection of it.
var jobType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Job",
	Fields: graphql.Fields{
		"id":     &graphql.Field{Type: graphql.String},
		"kind":   &graphql.Field{Type: graphql.String},
		"status": &graphql.Field{Type: graphql.String},
	},
})

// jobLogType is a schema-shape placeholder matching job_logs — see jobType.
var jobLogType = graphql.NewObject(graphql.ObjectConfig{
	Name: "JobLog",
	Fields: graphql.Fields{
		"loggedAt": &graphql.Field{Type: graphql.String},
		"level":    &graphql.Field{Type: graphql.String},
		"message":  &graphql.Field{Type: graphql.String},
	},
})

// jobsGap is the flagged reason every Jobs field below fails: there is no
// Jobs application service to call, and a resolver must not reach around one
// by querying the `jobs` table directly. Building a real Jobs system is out
// of scope for this slice; the schema shape exists so the surface is visible,
// but every resolver is a stub.
const jobsGap = "jobs infrastructure is not implemented yet (migration 0006_jobs.sql created tables only; no JobStore contract or application service exists)"

// jobsField stubs the "job list" query — see jobsGap.
func jobsField() *graphql.Field {
	return notImplementedField(graphql.NewList(jobType), graphql.FieldConfigArgument{
		"callerSessionId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
	}, jobsGap)
}

// jobField stubs the "job detail" query — see jobsGap.
func jobField() *graphql.Field {
	return notImplementedField(jobType, graphql.FieldConfigArgument{
		"callerSessionId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
		"id":              &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
	}, jobsGap)
}

// jobLogsField stubs the "job logs" query — see jobsGap.
func jobLogsField() *graphql.Field {
	return notImplementedField(graphql.NewList(jobLogType), graphql.FieldConfigArgument{
		"callerSessionId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
		"jobId":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
	}, jobsGap)
}

// retryJobField stubs the "retry command" mutation — see jobsGap.
func retryJobField() *graphql.Field {
	return notImplementedField(jobType, graphql.FieldConfigArgument{
		"callerSessionId": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
		"jobId":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
	}, jobsGap)
}
