package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"ecommerce/notification-service/repository"
	"ecommerce/shared/models"
)

type NotificationService struct {
	repo   *repository.NotificationRepository
	logger *zap.Logger
}

func NewNotificationService(repo *repository.NotificationRepository, logger *zap.Logger) *NotificationService {
	return &NotificationService{
		repo:   repo,
		logger: logger,
	}
}

// SendOrderConfirmation sends order confirmation notification
func (s *NotificationService) SendOrderConfirmation(userID, orderID string, totalPrice float64) error {
	s.logger.Info("Sending order confirmation",
		zap.String("user_id", userID),
		zap.String("order_id", orderID),
	)

	// Create notification record
	notification := &models.Notification{
		UserID:  userID,
		Type:    "email",
		Subject: "Order Confirmation",
		Message: fmt.Sprintf("Your order %s has been confirmed! Total: $%.2f", orderID, totalPrice),
		Status:  "pending",
	}

	// Save to database
	if err := s.repo.Create(context.Background(), notification); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	// Actually send notification (email, SMS, push, etc.)
	if err := s.sendNotification(notification); err != nil {
		s.logger.Error("Failed to send notification", zap.Error(err))
		// Mark as failed
		s.repo.UpdateStatus(context.Background(), notification.ID, "failed")
		return err
	}

	// Mark as sent
	s.repo.UpdateStatus(context.Background(), notification.ID, "sent")

	s.logger.Info("Notification sent successfully", zap.String("notification_id", notification.ID))
	return nil
}

// SendOrderCancellation sends order cancellation notification
func (s *NotificationService) SendOrderCancellation(userID, orderID string) error {
	s.logger.Info("Sending order cancellation",
		zap.String("user_id", userID),
		zap.String("order_id", orderID),
	)

	notification := &models.Notification{
		UserID:  userID,
		Type:    "email",
		Subject: "Order Cancelled",
		Message: fmt.Sprintf("Your order %s has been cancelled.", orderID),
		Status:  "pending",
	}

	if err := s.repo.Create(context.Background(), notification); err != nil {
		return fmt.Errorf("failed to save notification: %w", err)
	}

	if err := s.sendNotification(notification); err != nil {
		s.logger.Error("Failed to send notification", zap.Error(err))
		s.repo.UpdateStatus(context.Background(), notification.ID, "failed")
		return err
	}

	s.repo.UpdateStatus(context.Background(), notification.ID, "sent")
	return nil
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	return s.repo.UpdateStatus(ctx, notificationID, "read")
}

// sendNotification actually sends the notification via email/SMS/push
func (s *NotificationService) sendNotification(notification *models.Notification) error {
	// In production, integrate with:
	// - SendGrid/AWS SES for email
	// - Twilio for SMS
	// - Firebase for push notifications

	s.logger.Info("Sending notification",
		zap.String("type", notification.Type),
		zap.String("user_id", notification.UserID),
		zap.String("subject", notification.Subject),
	)

	// Simulate sending (in production, make actual API calls)
	// For now, just log
	s.logger.Info("Notification content", zap.String("message", notification.Message))

	return nil
}

// HealthCheck verifies service health
func (s *NotificationService) HealthCheck(ctx context.Context) error {
	return s.repo.HealthCheck(ctx)
}
