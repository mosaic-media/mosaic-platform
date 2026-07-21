// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package contracts

import "time"

// Clock provides a deterministic time boundary. Domain and application code
// must call Clock.Now instead of time.Now directly so tests can substitute a
// fixed clock.
type Clock interface {
	Now() time.Time
}
