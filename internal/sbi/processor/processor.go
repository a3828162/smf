package processor

import (
	"context"
	"sync"

	"github.com/free5gc/smf/internal/decisioner"
	"github.com/free5gc/smf/internal/logger"
	"github.com/free5gc/smf/internal/sbi/consumer"
	"github.com/free5gc/smf/pkg/app"
)

const (
	CONTEXT_NOT_FOUND = "CONTEXT_NOT_FOUND"
)

type ProcessorSmf interface {
	app.App

	Consumer() *consumer.Consumer
	CancelContext() context.Context
}

type Processor struct {
	ProcessorSmf
	decisioner.Decisioner

	roundRobin    int64
	roundRobinMax int64
	roundRobinMu  sync.Mutex

	randSlice []int64
	randMax   int64
	randIndex int64
}

func NewProcessor(smf ProcessorSmf) (*Processor, error) {
	p := &Processor{
		ProcessorSmf: smf,
	}
	decisioner, err := decisioner.NewDecisioner(smf.CancelContext(), smf.Config())

	if err != nil {
		logger.ProcessorLog.Errorf("NewDecisioner: %+v", err)
		return p, nil
	}
	if err = decisioner.Start(); err != nil {
		return p, err
	}
	p.Decisioner = decisioner

	p.roundRobin = 0
	p.roundRobinMax = 3
	p.roundRobinMu = sync.Mutex{}

	p.randSlice = []int64{0, 1, 1, 1, 2, 0}
	p.randMax = 6
	p.randIndex = 0

	return p, nil
}
