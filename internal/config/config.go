package config

import (
	"os"
)

type Config struct {
	TCP   TCPConfig
	HTTP  HTTPConfig
	MySQL MySQLConfig
	JWT   JWTConfig
	MQ    MQConfig
}

type MQConfig struct {
	URL       string
	QueueName string
}

type JWTConfig struct {
	Secret string
}

type TCPConfig struct {
	Host string
	Port int
}

type HTTPConfig struct {
	Host string
	Port int
}

type MySQLConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	DataName string
}

func Load() *Config {
	return &Config{
		TCP: TCPConfig{
			Host: envOr("TCP_HOST", "127.0.0.1"),
			Port: 8080,
		},
		HTTP: HTTPConfig{
			Host: envOr("HTTP_HOST", "127.0.0.1"),
			Port: 8081,
		},
		MySQL: MySQLConfig{
			Host:     envOr("MYSQL_HOST", "127.0.0.1"),
			Port:     3306,
			Username: envOr("MYSQL_USER", "root"),
			Password: os.Getenv("MYSQL_PASSWORD"),
			DataName: envOr("MYSQL_DATABASE", "im_system"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
		MQ: MQConfig{
			URL:       os.Getenv("RABBITMQ_URL"),
			QueueName: "im.message.persist",
		},
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
