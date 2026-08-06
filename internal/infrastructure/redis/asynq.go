package redis

import (
	"fmt"

	"github.com/hibiken/asynq"
)

func NewAsynqClient(redisURL string) (*asynq.Client, error) {
	opts, err := parseRedisURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := asynq.NewClient(opts)
	return client, nil
}

func NewAsynqServer(redisURL string, concurrency int) (*asynq.Server, error) {
	opts, err := parseRedisURL(redisURL)
	if err != nil {
		return nil, err
	}

	server := asynq.NewServer(opts, asynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})

	return server, nil
}

func NewAsynqScheduler(redisURL string) (*asynq.Scheduler, error) {
	opts, err := parseRedisURL(redisURL)
	if err != nil {
		return nil, err
	}

	scheduler := asynq.NewScheduler(opts, nil)
	return scheduler, nil
}

func parseRedisURL(redisURL string) (asynq.RedisClientOpt, error) {
	// Parse redis:// URL format
	// Format: redis://[:password@]host[:port][/db]
	
	var opt asynq.RedisClientOpt
	
	// Simple parsing for redis:// URLs
	if len(redisURL) < 8 {
		return opt, fmt.Errorf("invalid redis URL")
	}
	
	// Remove redis:// prefix
	url := redisURL[8:]
	
	// Check for password
	if idx := findIndex(url, '@'); idx != -1 {
		opt.Password = url[:idx]
		url = url[idx+1:]
	}
	
	// Split host and db
	if idx := findIndex(url, '/'); idx != -1 {
		host := url[:idx]
		db := url[idx+1:]
		if db != "" {
			// Parse db number
			opt.DB = parseInt(db)
		}
		opt.Addr = host
	} else {
		opt.Addr = url
	}
	
	// Default port
	if findIndex(opt.Addr, ':') == -1 {
		opt.Addr = opt.Addr + ":6379"
	}
	
	return opt, nil
}

func findIndex(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
