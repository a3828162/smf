package flask

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/smf/internal/logger"
	"github.com/free5gc/smf/pkg/factory"
	smf_model "github.com/free5gc/smf/pkg/models"
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

func (ds *DecisionerFlask) GetDecision(ctx context.Context, data1 any, data2 any, data3 any) (any, error) {
	logger.DecisionLog.Infoln("Get Decision from Flask Decision")

	postData := make(map[string]interface{})

	// 處理 DnPerf - 接收指標類型
	dnPerfPtr, ok := data1.(*models.DnPerfInfo)
	if !ok || dnPerfPtr == nil {
		logger.DecisionLog.Errorf("Data1 type assertion failed, expected *models.DnPerfInfo")
		return nil, fmt.Errorf("Data1 type assertion failed")
	}
	postData["dnPerf"] = *dnPerfPtr

	// Log DnPerfInfo 詳細資料
	logger.DecisionLog.Infof("========== DnPerf Info Details ==========")
	logger.DecisionLog.Infof("AppId: %s", dnPerfPtr.AppId)
	logger.DecisionLog.Infof("Dnn: %s", dnPerfPtr.Dnn)
	if dnPerfPtr.Snssai != nil {
		logger.DecisionLog.Infof("Snssai - SST: %d, SD: %s", dnPerfPtr.Snssai.Sst, dnPerfPtr.Snssai.Sd)
	} else {
		logger.DecisionLog.Infof("Snssai: nil")
	}
	logger.DecisionLog.Infof("Confidence: %d", dnPerfPtr.Confidence)

	// Log 每個 DnPerf 的 PerfData
	if len(dnPerfPtr.DnPerf) > 0 {
		logger.DecisionLog.Infof("DnPerf array has %d performance entries:", len(dnPerfPtr.DnPerf))
		for i, perf := range dnPerfPtr.DnPerf {
			logger.DecisionLog.Infof("  [%d] DNAI: %s", i, perf.Dnai)
			if perf.PerfData != nil {
				logger.DecisionLog.Infof("      Traffic Rate - Avg: %s, Max: %s",
					perf.PerfData.AvgTrafficRate, perf.PerfData.MaxTrafficRate)
				logger.DecisionLog.Infof("      Packet Delay - Avg: %d ms, Max: %d ms",
					perf.PerfData.AvePacketDelay, perf.PerfData.MaxPacketDelay)
				logger.DecisionLog.Infof("      Packet Loss Rate - Avg: %d (0.1%%)",
					perf.PerfData.AvgPacketLossRate)
				logger.DecisionLog.Infof("      Number of UEs: %d", perf.PerfData.NumOfUe)
			} else {
				logger.DecisionLog.Warnf("      PerfData is nil for DNAI: %s", perf.Dnai)
			}
		}
	} else {
		logger.DecisionLog.Warnf("DnPerf array is empty")
	}
	logger.DecisionLog.Infof("=========================================")

	// 處理 EdgeResource - 接收指標類型
	edgeResourcePtr, ok := data2.(*[]smf_model.EdgeMetric)
	if !ok || edgeResourcePtr == nil {
		logger.DecisionLog.Errorf("Data2 type assertion failed, expected *[]smf_model.EdgeMetric")
		return nil, fmt.Errorf("Data2 type assertion failed")
	}
	postData["edgeResource"] = *edgeResourcePtr
	logger.DecisionLog.Infof("Edge Resource Info: %+v", *edgeResourcePtr)

	// 處理 DNAI - 接收指標類型，注意 Dnais 是 []*DnaiInfo
	dnaisPtr, ok := data3.(*[]*factory.DnaiInfo)
	if !ok || dnaisPtr == nil {
		logger.DecisionLog.Errorf("Data3 type assertion failed, expected *[]*factory.DnaiInfo")
		return nil, fmt.Errorf("Data3 type assertion failed")
	}
	postData["dnais"] = *dnaisPtr
	logger.DecisionLog.Infof("DNAI Info: %+v", *dnaisPtr)

	logger.DecisionLog.Infof("Post Data: %+v", postData)

	var err2 error

	jsonData, err2 := json.Marshal(postData)
	if err2 != nil {
		retErr := fmt.Errorf("Failed to marshal data: %+v", err2)
		logger.DecisionLog.Errorf("%+v", retErr)
		return nil, retErr
	}

	client := ds.Client
	requestURL := ds.Endpoint + "/decisioner/select"

	logger.DecisionLog.Infof("Request URL: %s", requestURL)

	req, errReq := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(jsonData))
	if errReq != nil {
		logger.DecisionLog.Errorf("Failed to create request: %+v", errReq)
		return nil, errReq
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err2 := client.Do(req)
	if err2 != nil || resp == nil {
		logger.DecisionLog.Errorf("Failed to send request: %+v", err2)
		return nil, err2
	}

	defer func() {
		if err2 = resp.Body.Close(); err2 != nil {
			err2 = fmt.Errorf("Failed to close response body %+v", err2)
			logger.DecisionLog.Error(err2.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err2 = fmt.Errorf("GetPredict() response error: %d", resp.StatusCode)
		logger.DecisionLog.Error(err2.Error())
		return nil, err2
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.DecisionLog.Errorf("Failed to read response body: %+v", err)
		return nil, err
	}

	var result smf_model.APIResponse
	if err = json.Unmarshal(body, &result); err != nil {
		logger.DecisionLog.Errorf("Failed to unmarshal response body: %+v", err)
		return nil, err
	}
	logger.DecisionLog.Infof("GetDecision() success: %+v", result)

	return result, nil
}

func (ds *DecisionerFlask) BuildPrompt() interface{} {

	return map[string]interface{}{}
}
