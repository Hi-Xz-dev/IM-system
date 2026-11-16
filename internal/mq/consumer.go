package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"IM-system/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewConsumer(url, queueName string) (*Consumer, error) {
	if queueName == "" {
		return nil, fmt.Errorf("queue name is required")
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect consumer: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open consumer channel: %w", err)
	}

	//不是再创建一个不同的队列，而是确保同名队列存在、配置一致。
	_, err = ch.QueueDeclare(
		queueName,
		true,  // 队列持久化
		false, // 不自动删除
		false, // 不独占
		false, // 等待声明结果
		amqp.Table{
			"x-queue-type": "quorum",
		},
	)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("declare consumer queue: %w", err)
	}

	//1：每个消费者最多持有一条尚未 ACK 的消息。
	//0：不设置按字节大小计算的预取限制。
	//false：数量限制应用到每个消费者。
	if err := ch.Qos(1, 0, false); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("set consumer qos: %w", err)
	}

	return &Consumer{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}, nil
}

func (c *Consumer) Close() error {
	return c.conn.Close()
}

func (c *Consumer) Consume(
	ctx context.Context,
) (<-chan amqp.Delivery, error) {
	deliveries, err := c.channel.ConsumeWithContext(
		ctx,
		c.queueName,
		"",    // 让 RabbitMQ 生成消费者标识
		false, // autoAck：关闭自动确认，后面手动 ACK
		false, // exclusive：不独占队列
		false, // noLocal：这里保持 false
		false, // noWait：等待注册结果
		nil,   // 额外参数
	)
	if err != nil {
		return nil, fmt.Errorf("start consumer: %w", err)
	}

	return deliveries, nil
}

func (c *Consumer) Run(
	ctx context.Context,
	persist func(context.Context, domain.Message) error,
) error {
	// 本次消费结束时关闭连接，未 ACK 的消息可重新投递。
	defer c.Close()

	deliveries, err := c.Consume(ctx)
	if err != nil {
		return err
	}

	//取出消息 → JSON 转成 Message → 调用 Persist → 成功后 ACK
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case delivery, ok := <-deliveries:
			if !ok {
				if err := ctx.Err(); err != nil {
					return err
				}
				return fmt.Errorf("consumer delivery channel closed")
			}

			var msg domain.Message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				return fmt.Errorf("decode message: %w", err)
			}

			if err := persist(ctx, msg); err != nil {
				return fmt.Errorf("persist consumed message: %w", err)
			}

			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack consumed message: %w", err)
			}
		}
	}
}
