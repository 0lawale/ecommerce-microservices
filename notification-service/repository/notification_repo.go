package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"ecommerce/shared/models"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	notification.ID = uuid.New().String()
	notification.CreatedAt = time.Now()

	query := `
		INSERT INTO notifications (id, user_id, type, subject, message, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		notification.ID, notification.UserID, notification.Type,
		notification.Subject, notification.Message, notification.Status,
		notification.CreatedAt,
	)
	return err
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	query := `
		SELECT id, user_id, type, subject, message, status, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Subject, &n.Message, &n.Status, &n.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}
	return notifications, nil
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE notifications SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *NotificationRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
