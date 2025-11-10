package messaging

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"ecommerce/notification-service/service"
)

// OrderEvent represents an order event from the queue
type OrderEvent struct {
	OrderID    string    `json:"order_id"`
	UserID     string    `json:"user_id"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// RabbitMQConsumer consumes messages from RabbitMQ
type RabbitMQConsumer struct {
	conn                *amqp.Connection
	channel             *amqp.Channel
	notificationService *service.NotificationService
	logger              *zap.Logger
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(url string, notificationService *service.NotificationService, logger *zap.Logger) (*RabbitMQConsumer, error) {
	// Connect to RabbitMQ
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange (must match publisher's exchange)
	err = channel.ExchangeDeclare(
		"orders", // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue for this service
	queue, err := channel.QueueDeclare(
		"notifications", // name
		true,            // durable (survives broker restart)
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name, // queue name
		"",         // routing key (ignored for fanout)
		"orders",   // exchange
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	// Set QoS - process one message at a time
	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	logger.Info("RabbitMQ consumer initialized", zap.String("queue", queue.Name))

	return &RabbitMQConsumer{
		conn:                conn,
		channel:             channel,
		notificationService: notificationService,
		logger:              logger,
	}, nil
}

// StartConsuming starts consuming messages from the queue
func (c *RabbitMQConsumer) StartConsuming() error {
	// Register consumer
	messages, err := c.channel.Consume(
		"notifications",        // queue
		"notification-service", // consumer tag
		false,                  // auto-ack (we'll manually ack after processing)
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	c.logger.Info("Waiting for messages...")

	// Process messages
	for msg := range messages {
		c.processMessage(msg)
	}

	return nil
}

// processMessage processes a single message
func (c *RabbitMQConsumer) processMessage(msg amqp.Delivery) {
	c.logger.Info("Received message",
		zap.String("body", string(msg.Body)),
		zap.Time("timestamp", msg.Timestamp),
	)

	// Parse message
	var event OrderEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.logger.Error("Failed to parse message", zap.Error(err))
		// Reject message (won't be requeued)
		msg.Nack(false, false)
		return
	}

	// Process event based on status
	var err error
	switch event.Status {
	case "confirmed":
		err = c.notificationService.SendOrderConfirmation(event.UserID, event.OrderID, event.TotalPrice)
	case "cancelled":
		err = c.notificationService.SendOrderCancellation(event.UserID, event.OrderID)
	default:
		c.logger.Warn("Unknown order status", zap.String("status", event.Status))
	}

	// Acknowledge or reject message
	if err != nil {
		c.logger.Error("Failed to process message", zap.Error(err))
		// Nack with requeue - will retry later
		msg.Nack(false, true)
	} else {
		c.logger.Info("Message processed successfully", zap.String("order_id", event.OrderID))
		// Acknowledge message
		msg.Ack(false)
	}
}

// Close closes the RabbitMQ connection
func (c *RabbitMQConsumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
