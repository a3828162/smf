package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/free5gc/openapi/models"
	Nnwdaf_AnalyticsInfo "github.com/free5gc/openapi/nwdaf/AnalyticsInfo"
	"github.com/free5gc/smf/internal/logger"
	smf_model "github.com/free5gc/smf/pkg/models"
)

type nnwdafService struct {
	consumer *Consumer

	analyticsInfoMu   sync.RWMutex
	edgeResouceInfoMu sync.RWMutex

	analyticsInfoClients   map[string]*Nnwdaf_AnalyticsInfo.APIClient
	edgeResouceInfoClients *http.Client
}

// func (s *nnwdafService) searchNWDAF() string {
// 	searchResult, err := s.consumer.SendSearchNFInstances(
// 		s.consumer.Context().NrfUri, models.NrfNfManagementNfType_NWDAF)
// 	if err != nil {
// 		logger.ConsumerLog.Errorf("searchUDM error: %+v", err)
// 		return ""
// 	}

// 	logger.ConsumerLog.Infof("Search NWDAF Result: %+v", searchResult.NfInstanceList)
// 	url := ""
// 	for _, nfInstance := range searchResult.NfInstanceList {
// 		url = nfInstance.NrfDiscApiUri
// 		break
// 	}
// 	logger.ConsumerLog.Infof("NWDAF Uri: %+v", url)
// 	return url
// }

func (s *nnwdafService) getAnalyticsInfoClient(uri string) *Nnwdaf_AnalyticsInfo.APIClient {
	if uri == "" {
		return nil
	}
	s.analyticsInfoMu.RLock()
	client, ok := s.analyticsInfoClients[uri]
	if ok {
		s.analyticsInfoMu.RUnlock()
		return client
	}

	configuration := Nnwdaf_AnalyticsInfo.NewConfiguration()
	configuration.SetBasePath(uri)
	client = Nnwdaf_AnalyticsInfo.NewAPIClient(configuration)

	s.analyticsInfoMu.RUnlock()
	s.analyticsInfoMu.Lock()
	defer s.analyticsInfoMu.Unlock()
	s.analyticsInfoClients[uri] = client
	return client
}

func (s *nnwdafService) GetEdgeResouceInfo(ctx context.Context, nfUrl string) (
	*smf_model.EdcfEdgeResourceDataResponse,
	error,
) {
	reqCtx := context.WithoutCancel(ctx)
	requestUrl := fmt.Sprintf("%s/%s/%s", nfUrl, "nnwdaf-oam/v1", "getEdgeResourceData")

	logger.ConsumerLog.Debugf("GetEdgeResource() Request URL: %s", requestUrl)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, requestUrl, nil)
	if err != nil {
		logger.ConsumerLog.Errorf("NewRequestWithContext Error: %+v", err)
		return nil, err
	}
	// TODO
	// if s.consumer.Context().OAuth2Required {
	// GetTokenCtx() & BindToken()
	// }

	resp, err := s.edgeResouceInfoClients.Do(req)
	if err != nil {
		logger.ConsumerLog.Errorf("Do request error: %+v", err)
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.ConsumerLog.Errorf("Close response body error: %+v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Response status code: %d", resp.StatusCode)
	}

	var edgeResouce smf_model.EdcfEdgeResourceDataResponse
	if errJson := json.NewDecoder(resp.Body).Decode(&edgeResouce); errJson != nil {
		logger.ConsumerLog.Errorf("json Unmarshal error: %+v", errJson)
		return nil, errJson
	}
	logger.ConsumerLog.Debugf("GetEdgeResource() Response: %+v", edgeResouce)
	return &edgeResouce, nil
}

func (s *nnwdafService) GetNwdafAnalytics() (
	*models.NwdafAnalyticsInfoAnalyticsData,
	error,
) {
	logger.ConsumerLog.Infof("GetNwdafAnalytics-In Consumer")
	// uri := s.searchNWDAF()
	// example
	// client := s.getAnalyticsInfoClient(uri)
	client := s.getAnalyticsInfoClient("http://127.0.0.163:8000")

	// ctx, _, err := s.consumer.Context().GetTokenCtx(models.ServiceName_NNWDAF_ANALYTICSINFO, models.NrfNfManagementNfType_NWDAF)
	// if err != nil {
	// 	return nil, err
	// }

	request := Nnwdaf_AnalyticsInfo.GetNWDAFAnalyticsRequest{}
	request.SetEventId(models.EventId_DN_PERFORMANCE)

	// response, _, err :=
	res, err := client.NWDAFAnalyticsDocumentApi.GetNWDAFAnalytics(context.Background(), &request)
	if err != nil {
		logger.ConsumerLog.Errorf("Get NWDAF Analytics failed: %v", err)
		return nil, err
	}

	result := res.NwdafAnalyticsInfoAnalyticsData

	if result.DnPerfInfos != nil {
		logger.ConsumerLog.Debug("------------ DnPerfInfos Details ------------")
		for i, info := range result.DnPerfInfos {
			logger.ConsumerLog.Debugf("DnPerfInfo[%d]:", i)
			logger.ConsumerLog.Debugf("  AppId: %v", info.AppId)
			// if info.Snssai != nil {
			// 	logger.ConsumerLog.Debugf("  Snssai.Sst: %v", info.Snssai.Sst)
			// 	logger.ConsumerLog.Debugf("  Snssai.Sd: %v", info.Snssai.Sd)
			// }

			for i, perf := range info.DnPerf {
				logger.ConsumerLog.Debugf(" DnPerf[%d]:", i)
				logger.ConsumerLog.Debugf(" TerporalValidCon: %v", perf.TemporalValidCon)
				logger.ConsumerLog.Debugf(" Dnai: %v", perf.Dnai)
				logger.ConsumerLog.Debugf(" AvgTrafficRate: %v", perf.PerfData.AvgTrafficRate)
				logger.ConsumerLog.Debugf(" MaxTrafficRate: %v", perf.PerfData.MaxTrafficRate)
				logger.ConsumerLog.Debugf(" MinTrafficRate: %v", perf.PerfData.MinTrafficRate)
				logger.ConsumerLog.Debugf(" AggTrafficRate: %v", perf.PerfData.AggTrafficRate)
				logger.ConsumerLog.Debugf(" VarTrafficRate: %v", perf.PerfData.VarTrafficRate)
				// logger.ConsumerLog.Debugf(" TrafRateUeIds: %v", perf.PerfData.TrafRateUeIds)
				logger.ConsumerLog.Debugf(" AvePacketDelay: %v", perf.PerfData.AvePacketDelay)
				logger.ConsumerLog.Debugf(" MaxPacketDelay: %v", perf.PerfData.MaxPacketDelay)
				logger.ConsumerLog.Debugf(" VarPacketDelay: %v", perf.PerfData.VarPacketDelay)
				// logger.ConsumerLog.Debugf(" PackDelayUeIds: %v", perf.PerfData.PackDelayUeIds)
				logger.ConsumerLog.Debugf(" AvgPacketLossRate: %v", perf.PerfData.AvgPacketLossRate)
				logger.ConsumerLog.Debugf(" MaxPacketLossRate: %v", perf.PerfData.MaxPacketLossRate)
				logger.ConsumerLog.Debugf(" VarPacketLossRate: %v", perf.PerfData.VarPacketLossRate)
				// logger.ConsumerLog.Debugf(" PackLossUeIds: %v", perf.PerfData.PackLossUeIds)
				logger.ConsumerLog.Debugf(" NumOfUe: %v", perf.PerfData.NumOfUe)
				//
				// logger.ConsumerLog.Debugf(" PerfData: %v", perf.PerfData)
			}

		}
	}

	// logger.ConsumerLog.Infof("Get NWDAF Analytics response: %+v", result.DnPerfInfos)
	// use client to call NWDAF Analytics Info API
	// e.g., client.NWDAFAnalyticsDocumentApi.SomeAPIMethod(...)
	return &result, nil
}
