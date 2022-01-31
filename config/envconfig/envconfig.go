package envconfig

import (
	"os"
	"strconv"
)

type Config struct {
	LiveTimeout     int
	UpcomingTimeout int
	LiveURL         string
	UpcomingURL     string
	LiveJSON        string
	UpcomingJSON    string
	Addr            string
}

func NewConfig() *Config {
	return &Config{
		LiveTimeout:     getEnvInt("LIVE_TIMEOUT", 3),
		UpcomingTimeout: getEnvInt("UPCOMING_TIMEOUT", 20),
		LiveURL: getEnvString("LIVE_URL",
			"https://odds.stagbet.site/v1/events/40/0/list/300/live/ru"),
		UpcomingURL: getEnvString("UPCOMING_URL",
			"https://odds.stagbet.site/v1/events/40/0/list/1000/line/ru"),
		LiveJSON:     getEnvString("LIVE_JSON", "live"),
		UpcomingJSON: getEnvString("UPCOMING_JSON", "upcoming"),
		Addr:         getEnvString("WS_ADDR", ":8081"),
	}
}

func getEnvString(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvInt(name string, defaultVal int) int {
	valueStr := getEnvString(name, strconv.Itoa(defaultVal))
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
