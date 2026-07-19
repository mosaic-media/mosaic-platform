package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mosaic-media/mosaic-platform/internal/platform/contracts"
	"github.com/mosaic-media/mosaic-platform/internal/platform/domain"
)

// relationStore is the PostgreSQL contracts.RelationStore. Reading a computed
// grouping back is an indexed join here, on the same engine as everything
// else, rather than a second query path.
type relationStore struct {
	q queryer
}

// NewRelationStore builds a pool-backed RelationStore for the direct (non-
// transactional) read path.
func NewRelationStore(pool *pgxpool.Pool) contracts.RelationStore {
	return &relationStore{q: pool}
}

const relationColumns = `id, from_node_id, to_node_id, relation_type, confidence, origin, created_at`

func (s *relationStore) Create(ctx context.Context, relation domain.Relation) (domain.Relation, error) {
	_, err := s.q.Exec(ctx,
		`INSERT INTO relations (id, from_node_id, to_node_id, relation_type, confidence, origin, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		string(relation.ID), string(relation.FromNodeID), string(relation.ToNodeID),
		string(relation.Type), relation.Confidence, string(relation.Origin), relation.CreatedAt,
	)
	if err != nil {
		return domain.Relation{}, mapError("create relation", err)
	}
	return relation, nil
}

func (s *relationStore) FindByID(ctx context.Context, id domain.RelationID) (domain.Relation, error) {
	row := s.q.QueryRow(ctx, `SELECT `+relationColumns+` FROM relations WHERE id = $1`, string(id))
	relation, err := scanRelation(row)
	if err != nil {
		if isNoRows(err) {
			return domain.Relation{}, contracts.NewError(contracts.NotFound, "relation not found")
		}
		return domain.Relation{}, mapError("find relation by id", err)
	}
	return relation, nil
}

func (s *relationStore) ListFrom(ctx context.Context, from domain.NodeID, relationType domain.RelationType) ([]domain.Relation, error) {
	rows, err := s.q.Query(ctx,
		`SELECT `+relationColumns+` FROM relations
		 WHERE from_node_id = $1 AND ($2 = '' OR relation_type = $2)
		 ORDER BY relation_type, created_at, id`,
		string(from), string(relationType),
	)
	if err != nil {
		return nil, mapError("list relations from node", err)
	}
	return collectRelations(rows, "list relations from node")
}

func (s *relationStore) ListTo(ctx context.Context, to domain.NodeID, relationType domain.RelationType) ([]domain.Relation, error) {
	rows, err := s.q.Query(ctx,
		`SELECT `+relationColumns+` FROM relations
		 WHERE to_node_id = $1 AND ($2 = '' OR relation_type = $2)
		 ORDER BY relation_type, created_at, id`,
		string(to), string(relationType),
	)
	if err != nil {
		return nil, mapError("list relations to node", err)
	}
	return collectRelations(rows, "list relations to node")
}

func (s *relationStore) Delete(ctx context.Context, id domain.RelationID) error {
	tag, err := s.q.Exec(ctx, `DELETE FROM relations WHERE id = $1`, string(id))
	if err != nil {
		return mapError("delete relation", err)
	}
	if tag.RowsAffected() == 0 {
		return contracts.NewError(contracts.NotFound, "relation not found")
	}
	return nil
}

func scanRelation(row pgx.Row) (domain.Relation, error) {
	var (
		relation     domain.Relation
		id           string
		fromNodeID   string
		toNodeID     string
		relationType string
		origin       string
	)
	if err := row.Scan(&id, &fromNodeID, &toNodeID, &relationType,
		&relation.Confidence, &origin, &relation.CreatedAt); err != nil {
		return domain.Relation{}, err
	}
	relation.ID = domain.RelationID(id)
	relation.FromNodeID = domain.NodeID(fromNodeID)
	relation.ToNodeID = domain.NodeID(toNodeID)
	relation.Type = domain.RelationType(relationType)
	relation.Origin = domain.RelationOrigin(origin)
	return relation, nil
}

func collectRelations(rows pgx.Rows, message string) ([]domain.Relation, error) {
	defer rows.Close()

	var relations []domain.Relation
	for rows.Next() {
		relation, err := scanRelation(rows)
		if err != nil {
			return nil, mapError("scan relation row", err)
		}
		relations = append(relations, relation)
	}
	if err := rows.Err(); err != nil {
		return nil, mapError(message, err)
	}
	return relations, nil
}
