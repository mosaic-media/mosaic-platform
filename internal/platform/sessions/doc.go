// Package sessions holds session issuance, validation and revocation
// (MEG-015 §07). Sessions are server-issued and revocable: validation
// checks the persisted record on every call rather than trusting a
// client-held token, and revocation always writes through a SessionStore
// rather than relying on the client discarding anything.
package sessions
