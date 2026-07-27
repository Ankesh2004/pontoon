package worker

import (
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/redis"
)

type Processor struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewProcessor(redisURL string, concurrency int, deployProcessor *DeployProcessor) (*Processor, error) {
	server, err := redis.NewAsynqServer(redisURL, concurrency)
	if err != nil {
		return nil, err
	}

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeDeploy, deployProcessor.ProcessDeployTask)

	return &Processor{
		server: server,
		mux:    mux,
	}, nil
}

func (p *Processor) Start() error {
	slog.Info("starting worker processor")
	if err := p.server.Start(p.mux); err != nil {
		return err
	}
	return nil
}

func (p *Processor) Stop() {
	slog.Info("stopping worker processor")
	p.server.Stop()
}
