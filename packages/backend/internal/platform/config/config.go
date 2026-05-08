package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Server ServerConfig `mapstructure:"server"`
	Auth   AuthConfig   `mapstructure:"auth"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Influx InfluxConfig `mapstructure:"influx"`
	MQTT   MQTTConfig   `mapstructure:"mqtt"`
	Log    LogConfig    `mapstructure:"log"`
	Device DeviceConfig `mapstructure:"device"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type AuthConfig struct {
	JWTSecret       string `mapstructure:"jwt_secret"`
	TokenExpireSecs int    `mapstructure:"token_expire_secs"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Params   string `mapstructure:"params"`
}

type InfluxConfig struct {
	URL    string `mapstructure:"url"`
	Token  string `mapstructure:"token"`
	Org    string `mapstructure:"org"`
	Bucket string `mapstructure:"bucket"`
}

type MQTTConfig struct {
	Broker   string `mapstructure:"broker"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	ClientID string `mapstructure:"client_id"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type DeviceConfig struct {
	HeartbeatTimeoutSec int `mapstructure:"heartbeat_timeout_sec"`
}

func Load() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")

	viper.SetEnvPrefix("HAMB")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Auth.JWTSecret == "" || cfg.Auth.JWTSecret == "change-me" {
		return fmt.Errorf("auth.jwt_secret: must not be empty or default value \"change-me\" — generate a strong secret (>= 32 chars)")
	}
	if len(cfg.Auth.JWTSecret) < 32 {
		return fmt.Errorf("auth.jwt_secret: must be at least 32 characters (got %d)", len(cfg.Auth.JWTSecret))
	}

	var warnings []string
	if cfg.MySQL.Password == "root" || cfg.MySQL.Password == "" {
		warnings = append(warnings, "mysql.password: using default/empty password")
	}
	if cfg.Influx.Token == "your-token" || cfg.Influx.Token == "" {
		warnings = append(warnings, "influx.token: using default/empty token")
	}
	if cfg.MQTT.Password == "public" || cfg.MQTT.Password == "" {
		warnings = append(warnings, "mqtt.password: using default/empty password")
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "[CONFIG WARNING] %s\n", w)
	}

	return nil
}
