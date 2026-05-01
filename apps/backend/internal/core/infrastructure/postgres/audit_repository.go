package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"opensplit/apps/backend/internal/core/domain"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Save(ctx context.Context, log domain.AuditLog) error {
	query := `INSERT INTO audit_logs (id, group_id, user_id, action, target_id, details) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, log.ID, log.GroupID, log.UserID, log.Action, log.TargetID, log.Details)
	return err
}

func (r *AuditRepository) ListByGroup(ctx context.Context, groupID domain.GroupID) ([]domain.AuditLog, error) {
	query := `SELECT id, group_id, user_id, action, target_id, details, created_at 
	          FROM audit_logs WHERE group_id = $1 ORDER BY created_at DESC LIMIT 50`

	rows, err := r.db.QueryContext(ctx, query, string(groupID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(&l.ID, &l.GroupID, &l.UserID, &l.Action, &l.TargetID, &l.Details, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating rows: %w", err)
		}
	}
	return logs, nil
}
