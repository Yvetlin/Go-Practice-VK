// Author: Yvetlin [Gloria]
// Date: 2025
package config

import "os"

type Config struct {
	GRPCPort      string
	TestGRPCAddr1 string
	TestGRPCAddr2 string
}

func Load() *Config {
	return &Config{
		GRPCPort:      getEnv("GRPC_PORT", "50051"),
		TestGRPCAddr1: getEnv("TEST_GRPC_ADDR1", "localhost:50052"),
		TestGRPCAddr2: getEnv("TEST_GRPC_ADDR2", "localhost:50053"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
