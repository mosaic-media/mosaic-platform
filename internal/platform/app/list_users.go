// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package app

import (
	"context"

	"github.com/mosaic-media/platform/internal/platform/contracts"
	"github.com/mosaic-media/platform/internal/platform/domain"
	"github.com/mosaic-media/platform/internal/platform/policy"
)

// ActionUserList is the policy action evaluated for ListUsers.
const ActionUserList policy.Action = "user.list"

// ListUsersQuery lists every local Platform user.
type ListUsersQuery struct {
	CallerSessionID domain.SessionID
}

// ListUsersResult is the Platform result type returned by ListUsers.
type ListUsersResult struct {
	Users []domain.User
}

func validateListUsersQuery(query ListUsersQuery) error {
	if query.CallerSessionID == "" {
		return contracts.NewError(contracts.InvalidArgument, "caller session id is required")
	}
	return nil
}

// ListUsers implements the query boundary: authenticate and authorize before
// reading state, no UnitOfWork needed for a read.
func (s *Service) ListUsers(ctx context.Context, query ListUsersQuery) (ListUsersResult, error) {
	if err := validateListUsersQuery(query); err != nil {
		return ListUsersResult{}, err
	}

	if _, err := s.enterSession(ctx, query.CallerSessionID, ActionUserList,
		policy.Resource{Type: "user"}); err != nil {
		return ListUsersResult{}, err
	}

	users, err := s.users.List(ctx)
	if err != nil {
		return ListUsersResult{}, err
	}
	return ListUsersResult{Users: users}, nil
}
