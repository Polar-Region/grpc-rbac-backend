package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MysqlDsn      string
	AdminUsername string
	AdminPassword string
	Addr          string
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func Load() *Config {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
		log.Println("Using system environment variables or defaults")
	} else {
		log.Println("Successfully loaded .env file")
	}

	cfg := &Config{
		MysqlDsn:      getEnv("MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/testDB?charset=utf8mb4&parseTime=True&loc=Local"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "123456"),
		Addr:          getEnv("ADDR", ":8080"),
	}

	// 调试信息
	log.Printf("=== Configuration Loaded ===")
	log.Printf("MySQL DSN: %s", cfg.MysqlDsn)
	log.Printf("Admin Username: %s", cfg.AdminUsername)
	log.Printf("Admin Password: %s", cfg.AdminPassword)
	log.Printf("Address: %s", cfg.Addr)
	log.Printf("============================")

	return cfg
}
