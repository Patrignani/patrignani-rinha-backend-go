package config

type Environment struct {
	Postgres             Postgres
	StartPort            string `env:"START_PORT,default=9999"`
	ScreeningQueue       QueueScreening
	HighPriorityQueue    QueueHighPriority
	LowPriorityQueue     QueueLowPriority
	LowWaitingRoomQueue  QueueLowWaiting
	HighWaitingRoomQueue QueueHighWaiting
	KFactor              float32 `env:"K_FACTOR"`
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
	Buffer  int `env:"LOW_WAITING_BUFFER"`
	Workers int `env:"LOW_WAITING_WORKERS"`
}

type QueueHighWaiting struct {
	Buffer  int `env:"HIGH_WAITING_BUFFER"`
	Workers int `env:"HIGH_WAITING_WORKERS"`
}
