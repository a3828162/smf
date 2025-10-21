package flask

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/free5gc/smf/internal/logger"
	"github.com/free5gc/smf/pkg/factory"
)

type DecisionerFlask struct {
	Client   *http.Client
	ctx      context.Context
	Endpoint string
	// SupportNFSet map[models.NrfNfManagementNfType]bool
}

func NewDecisionerFlask(ctx context.Context, cfg *factory.Flask) (*DecisionerFlask, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Flask config is nil")
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("Endpoint is empty")
	}
	endpoint := cfg.Endpoint

	logger.DecisionLog.Debugf("New Flask Mtlf with endpoint: %s", endpoint)

	return &DecisionerFlask{
		Client:   &http.Client{},
		ctx:      ctx,
		Endpoint: endpoint,
	}, nil
}

func (ds *DecisionerFlask) Start() error {
	logger.DecisionLog.Infof("Start Flask Mtlf with endpoint: %s", ds.Endpoint)

	ds.healthCheck()
	return nil
}

func (ds *DecisionerFlask) healthCheck() {
	logger.DecisionLog.Debugln("Health check for Flask MTLF")

	client := ds.Client
	requestURL := ds.Endpoint + "/decisioner"

	req, errReq := http.NewRequestWithContext(context.Background(), http.MethodGet, requestURL, nil)
	if errReq != nil {
		logger.DecisionLog.Errorf("Failed to create request %+v", errReq)
		return
	}
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		logger.DecisionLog.Errorf("Health check failed: %+v", err)
		return
	}
	if resp.StatusCode != 200 {
		logger.DecisionLog.Errorf("Health check failed: %d", resp.StatusCode)
		return
	}
	if err = resp.Body.Close(); err != nil {
		logger.DecisionLog.Errorf("Failed to close response body %+v", err)
		return
	}
	logger.DecisionLog.Debugln("Health check success")
}

func (ds *DecisionerFlask) GetDecision(data1 any, data2 any) (any, error) {
	logger.DecisionLog.Infoln("Get Decision from Flask Decision")

	client := ds.Client
	requestURL := ds.Endpoint + "/decisioner/select"

	logger.DecisionLog.Debugf("Request URL: %s", requestURL)

	prompt := ds.BuildPrompt()
	jsonData, err := json.Marshal(prompt)
	if err != nil {
		retErr := fmt.Errorf("Failed to marshal data: %+v", err)
		logger.DecisionLog.Errorf("%+v", retErr)
		return nil, retErr
	}

	req, errReq := http.NewRequestWithContext(ds.ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonData))
	if errReq != nil {
		logger.DecisionLog.Errorf("Failed to create request: %+v", errReq)
		return nil, errReq
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp == nil {
		logger.DecisionLog.Errorf("Failed to send request: %+v", err)
		return nil, err
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			err = fmt.Errorf("Failed to close response body %+v", err)
			logger.DecisionLog.Error(err.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("GetPredict() response error: %d", resp.StatusCode)
		logger.DecisionLog.Error(err.Error())
		return nil, err
	}
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	logger.DecisionLog.Errorf("Failed to read response body: %+v", err)
	// 	return nil, err
	// }

	return nil, nil
}

func (ds *DecisionerFlask) BuildPrompt() interface{} {

	return map[string]interface{}{}
}
