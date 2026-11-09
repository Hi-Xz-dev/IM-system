package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"IM-system/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) Publish(ctx context.Context, msg domain.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("encode message: %w", err)
	}

	r.publishMu.Lock()
	defer r.publishMu.Unlock()

	// 如果等锁期间已经超时，就不再发送。
	if err := ctx.Err(); err != nil {
		return err
	}

	err = r.channel.PublishWithContext(
		ctx,
		"",          // 使用默认交换机  路由键等于队列名，就把消息交给那个队列
		r.queueName, // 路由键：这里用目标队列名
		true,        // mandatory：找不到队列时退回消息 无法路由时，通过之前注册的 returns 收到退回通知
		false,       // immediate：不使用

		amqp.Publishing{
			DeliveryMode: amqp.Persistent, //将消息标记为持久化消息，配合队列配置和发布确认使用
			ContentType:  "application/json",
			MessageId:    msg.ID,
			Body:         body, //真正发送的消息内容，即 JSON 字节
		},
	)
	if err != nil {
		_ = r.conn.Close()
		return fmt.Errorf("publish message: %w", err)
	}

	return r.waitPublishResult(ctx)
}

func (r *RabbitMQ) waitPublishResult(ctx context.Context) error {
	var returned *amqp.Return

	for {
		select {
		case ret, ok := <-r.returns:
			if !ok {
				return fmt.Errorf("publisher return channel closed")
			}
			returned = &ret // 记录退回结果，继续等待本次确认。

		case confirmation, ok := <-r.confirms:
			if !ok {
				return fmt.Errorf("publish outcome unknown: confirmation channel closed")
			}

			// 两种通知来自不同的 Go Channel。
			// select 可能先选中 ACK，所以再检查是否有退回通知。
			select {
			case ret, ok := <-r.returns:
				if ok {
					returned = &ret
				}
			default:
			}

			if returned != nil {
				return fmt.Errorf(
					"message was returned: %d %s",
					returned.ReplyCode,
					returned.ReplyText,
				)
			}

			if !confirmation.Ack {
				return fmt.Errorf("RabbitMQ rejected the message")
			}

			return nil

		case <-ctx.Done():
			// 防止迟到的确认被下一次发布误用。
			_ = r.conn.Close()
			return fmt.Errorf("publish outcome unknown: %w", ctx.Err())
		}
	}
}
