package config

type Config struct {
	TCP   TCPConfig
	HTTP  HTTPConfig
	MySQL MySQLConfig
	JWT   JWTConfig
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
			Host: "127.0.0.1",
			Port: 8080,
		},
		HTTP: HTTPConfig{
			Host: "127.0.0.1",
			Port: 8081,
		},
		MySQL: MySQLConfig{
			Host:     "127.0.0.1",
			Port:     3306,
			Username: "root",
			DataName: "im_system",
		},
		JWT: JWTConfig{
			Secret: "xx",
		},
	}
}
