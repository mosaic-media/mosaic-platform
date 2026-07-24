// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package extension

import v1 "github.com/mosaic-media/sdk/contracts/platform/v1"

// ResolvingContentForTest exposes the inbound half of the handle mechanism to
// the external test package, bound to a launched module's invocation table.
//
// It exists so a test can present a handle the Platform never minted — which is
// the case that matters and which cannot be produced through the module, since
// the module is only ever handed live handles. Testing the wrapper directly is
// testing the real path: this is the exact value the module's callbacks land on.
func ResolvingContentForTest(inner v1.ContentService, m *Module) v1.ContentService {
	return &resolvingContent{inner: inner, inv: m.invocations}
}
