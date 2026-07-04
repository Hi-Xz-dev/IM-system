package config

type Config struct {
	TCP  TCPConfig
	HTTP HTTPConfig
}

type TCPConfig struct {
	Host string
	Port int
}

type HTTPConfig struct {
	Host string
	Port int
}

func Load() *Config {
	return &Config{
		TCP: TCPConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		HTTP: HTTPConfig{
			Host: "127.0.0.1",
			Port: 8081,
		},
	}
}
