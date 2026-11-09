package mq

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
	confirms  <-chan amqp.Confirmation
	returns   <-chan amqp.Return
	publishMu sync.Mutex
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	//RabbitMQ 对发布消息给出 ACK 或 NACK
	confirms := ch.NotifyPublish(
		make(chan amqp.Confirmation, 1),
	)

	//发布时启用 mandatory 后，消息找不到目标队列，会收到退回通知。
	returns := ch.NotifyReturn(
		make(chan amqp.Return, 1),
	)
	q, err := ch.QueueDeclare(
		"im.message.persist", // 队列名称
		true,                 // durable：持久化队列
		false,                // autoDelete：不用时不自动删除
		false,                // exclusive：不独占，允许其他连接使用
		false,                // noWait：等待声明结果
		amqp.Table{
			"x-queue-type": "quorum", // 使用 quorum 队列
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn:      conn,
		channel:   ch,
		queueName: q.Name,
		confirms:  confirms,
		returns:   returns,
	}, nil
}

func (r *RabbitMQ) Close() error {
	return r.conn.Close()
}
