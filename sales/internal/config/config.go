package config

import "os"

func GetMRPURL() string {
	if v := os.Getenv("MRP_URL"); v != "" {
		return v
	}
	return "http://localhost:8082"
}

func GetSCMURL() string {
	if v := os.Getenv("SCM_URL"); v != "" {
		return v
	}
	return "http://localhost:8083"
}
