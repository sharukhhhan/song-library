package postgres

import (
	"context"
	"effective_mobile_tz/internal/entity"
	"effective_mobile_tz/internal/repository/repoerrors"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type GroupPostgres struct {
	*pgx.Conn
}

func NewGroupPostgres(conn *pgx.Conn) *GroupPostgres {
	return &GroupPostgres{Conn: conn}
}

func (g *GroupPostgres) CreateGroup(ctx context.Context, name string) (string, error) {
	query := `INSERT INTO groups (name) VALUES ($1) RETURNING id`
	var groupID string

	err := g.QueryRow(ctx, query, name).Scan(&groupID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return "", repoerrors.ErrAlreadyExists
			}
		}
		return "", err
	}

	return groupID, nil
}

func (g *GroupPostgres) GetGroupIDByName(ctx context.Context, name string) (string, error) {
	query := `SELECT id, name FROM groups WHERE name = $1`
	var group entity.Group

	err := g.QueryRow(ctx, query, name).Scan(&group.ID, &group.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repoerrors.ErrNotFound
		}
		return "", err
	}

	return group.ID, nil
}
