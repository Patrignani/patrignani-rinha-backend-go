package config

import "time"

type Environment struct {
	Postgres               Postgres
	StartPort              string `env:"START_PORT,default=8080"`
	ScreeningQueue         QueueScreening
	HighPriorityQueue      QueueHighPriority
	LowPriorityQueue       QueueLowPriority
	WaitingRoomQueue       QueueLowWaiting
	KFactor                float32       `env:"K_FACTOR"`
	LimitTimeHealth        int           `env:"LIMIT_TIME_HEALTH"`
	WaitingRoomSleepTime   time.Duration `env:"WAITING_ROOM_SLEEP_TIME"`
	DefaultUrl             string        `env:"DEFAULT_URL"`
	FallbackUrl            string        `env:"FALLBACK_URL"`
	EnableCheckHealthCheck bool          `env:"ENABLE_CHECK_HEALTH_CHECK"`
}

type Postgres struct {
	Host string `env:"DB_HOST"`
	User string `env:"DB_USER"`
	Pass string `env:"DB_PASSWORD"`
	Name string `env:"DB_NAME"`
	PORT string `env:"DB_PORT,default=5432"`
}

type QueueScreening struct {
	Buffer  int `env:"SCREENING_BUFFER"`
	Workers int `env:"SCREENING_WORKERS"`
}

type QueueHighPriority struct {
	Buffer  int `env:"HIGH_PRIORITY_BUFFER"`
	Workers int `env:"HIGH_PRIORITY_WORKERS"`
}

type QueueLowPriority struct {
	Buffer  int `env:"LOW_PRIORITY_BUFFER"`
	Workers int `env:"LOW_PRIORITY_WORKERS"`
}

type QueueLowWaiting struct {
	Buffer  int `env:"WAITING_BUFFER"`
	Workers int `env:"WAITING_WORKERS"`
}
