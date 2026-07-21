// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package domain

import "time"

// HealthState is the readiness state of a Platform component.
type HealthState string

const (
	HealthHealthy     HealthState = "healthy"
	HealthDegraded    HealthState = "degraded"
	HealthUnavailable HealthState = "unavailable"
)

// HealthStatus is a point-in-time readiness report for a Platform
// component. It is the simple shape contracts.HealthProbe returns;
// ComponentHealth below is the richer per-component diagnostics report the
// diagnostics model requires.
type HealthStatus struct {
	Component string
	State     HealthState
	Detail    string
	CheckedAt time.Time
}

// LifecycleState is a component's position in its own startup/operation
// lifecycle.
type LifecycleState string

const (
	LifecycleStarting LifecycleState = "starting"
	LifecycleRunning  LifecycleState = "running"
	LifecycleStopping LifecycleState = "stopping"
	LifecycleStopped  LifecycleState = "stopped"
)

// ComponentDependency names another component this component depends on,
// and that dependency's own current health — so a caller can tell why a
// component is degraded without separately walking the whole dependency
// graph itself.
type ComponentDependency struct {
	Component string
	Health    HealthState
}

// ComponentHealth is the full per-component diagnostics report the
// diagnostics model requires: component identifier, lifecycle state,
// health status, degraded reason, last successful check, last failure
// category, dependency health, and support bundle redaction class.
//
// DegradedReason (and any other free-text detail a reporter attaches) may
// describe internal error content; RedactionClass governs whether it may
// appear verbatim in a support bundle the same way it already governs
// domain.Event payloads. A
// reporter that has not classified itself should prefer RedactionSensitive
// over RedactionNone — the same redact-by-default bias RedactionClass
// documents for events.
type ComponentHealth struct {
	Component           string
	Lifecycle           LifecycleState
	Health              HealthState
	DegradedReason      string
	LastSuccessfulCheck time.Time
	LastFailureCategory string
	DependencyHealth    []ComponentDependency
	RedactionClass      RedactionClass
}
