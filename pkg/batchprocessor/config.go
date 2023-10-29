package batchprocessor

import (
	"log"
	"time"
)

type Config struct {
	RateLimitTimePeriod time.Duration
	RateLimitUnit       time.Duration
	logger              *log.Logger
}
