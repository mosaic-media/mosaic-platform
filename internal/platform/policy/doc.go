// Package policy holds the Platform's Policy Decision Point (MEG-015 §07,
// MEG-009 §04). It answers authorize(subject, action, resource, context)
// and returns an allow/deny Decision. The decision point may live
// in-process, as it does here; the enforcement point — the code that
// actually refuses to mutate state on a deny — belongs entirely to
// application services (internal/platform/app), not to this package.
package policy
