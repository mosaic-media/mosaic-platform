// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

// Package config defines Platform configuration schema, validation and activation.
package config

// Config is a placeholder configuration surface. It does not yet carry the
// real schema, reload-class tagging or activation versions.
type Config struct {
	Environment string
}

// Load returns a stub configuration. It does not yet read from disk, the
// environment or a config store.
func Load() (Config, error) {
	return Config{Environment: "local"}, nil
}
