package decisioner

import (
	"context"
	"fmt"

	"github.com/free5gc/smf/internal/decisioner/flask"
	"github.com/free5gc/smf/internal/logger"
	"github.com/free5gc/smf/pkg/factory"
)

const (
	Decisioner_TYPE_FLASK factory.MtlfType = "flask"
)

type Decisioner interface {
	Start() error
	GetDecision(any, any) (any, error)

	// // PostNewData() is a function to handle any data collected from NFs
	// // and process it to json format and send it to MTLF server
	// PostNewData(ctx context.Context, nfType models.NrfNfManagementNfType, data any) error
}

func NewDecisioner(ctx context.Context, cfg *factory.Config) (Decisioner, error) {
	decisionerType := cfg.Configuration.MtlfType
	if decisionerType == "" {
		return nil, fmt.Errorf("Not specify Decisioner")
	}
	logger.DecisionLog.Infof("New Mtlf: %s", decisionerType)

	if decisionerType == Decisioner_TYPE_FLASK {
		return flask.NewDecisionerFlask(ctx, cfg.Configuration.Flask)
	}
	return nil, fmt.Errorf("Unsupported Decisioner type: %s", decisionerType)
}
