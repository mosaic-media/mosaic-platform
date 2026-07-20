// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package contracts

import (
	"context"

	"github.com/mosaic-media/mosaic-platform/internal/platform/domain"
)

// SecretBroker provides secret retrieval and rotation (MEG-015 §03).
type SecretBroker interface {
	Resolve(ctx context.Context, ref domain.SecretRef) (domain.Secret, error)
	Rotate(ctx context.Context, ref domain.SecretRef) (domain.Secret, error)
}
