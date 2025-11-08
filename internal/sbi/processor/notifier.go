package processor

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/openapi/smf/EventExposure"
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/internal/logger"
	"github.com/free5gc/smf/pkg/factory"
	smf_model "github.com/free5gc/smf/pkg/models"
)

func (p *Processor) HandleChargingNotification(
	c *gin.Context,
	chargingNotifyRequest models.ChargingNotifyRequest,
	smContextRef string,
) {
	logger.ChargingLog.Info("Handle Charging Notification")

	problemDetails := p.chargingNotificationProcedure(chargingNotifyRequest, smContextRef)
	if problemDetails == nil {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(int(problemDetails.Status), problemDetails)
}

// While receive Charging Notification from CHF, SMF will send Charging Information to CHF and update UPF
// The Charging Notification will be sent when CHF found the changes of the quota file.
func (p *Processor) chargingNotificationProcedure(
	req models.ChargingNotifyRequest, smContextRef string,
) *models.ProblemDetails {
	if smContext := smf_context.GetSMContextByRef(smContextRef); smContext != nil {
		smContext.SMLock.Lock()
		defer smContext.SMLock.Unlock()
		upfUrrMap := make(map[string][]*smf_context.URR)
		for _, reauthorizeDetail := range req.ReauthorizationDetails {
			rg := reauthorizeDetail.RatingGroup
			logger.ChargingLog.Infof("Force update charging information for rating group %d", rg)
			for _, urr := range smContext.UrrUpfMap {
				chgInfo := smContext.ChargingInfo[urr.URRID]
				if chgInfo.RatingGroup == rg ||
					chgInfo.ChargingLevel == smf_context.PduSessionCharging {
					logger.ChargingLog.Tracef("Query URR (%d) for Rating Group (%d)", urr.URRID, rg)
					upfId := smContext.ChargingInfo[urr.URRID].UpfId
					upfUrrMap[upfId] = append(upfUrrMap[upfId], urr)
				}
			}
		}
		for upfId, urrList := range upfUrrMap {
			upf := smf_context.GetUpfById(upfId)
			if upf == nil {
				logger.ChargingLog.Warnf("Cound not find upf %s", upfId)
				continue
			}
			QueryReport(smContext, upf, urrList, models.ChfConvergedChargingTriggerType_FORCED_REAUTHORISATION)
		}
		p.ReportUsageAndUpdateQuota(smContext)
	} else {
		detail := fmt.Sprintf("SM Context [%s] Not Found ", smContextRef)
		return openapi.ProblemDetailsDataNotFound(detail)
	}

	return nil
}

func (p *Processor) HandleSMPolicyUpdateNotify(
	c *gin.Context,
	request models.SmPolicyNotification,
	smContextRef string,
) {
	logger.PduSessLog.Infoln("In HandleSMPolicyUpdateNotify")
	decision := request.SmPolicyDecision
	smContext := smf_context.GetSMContextByRef(smContextRef)

	if smContext == nil {
		logger.PduSessLog.Errorf("SMContext[%s] not found", smContextRef)
		c.Status(http.StatusBadRequest)
		return
	}

	smContext.SMLock.Lock()
	defer smContext.SMLock.Unlock()

	smContext.CheckState(smf_context.Active)
	// Wait till the state becomes Active again
	// TODO: implement waiting in concurrent architecture

	smContext.SetState(smf_context.ModificationPending)

	// Update SessionRule from decision
	if err := smContext.ApplySessionRules(decision); err != nil {
		// TODO: Fill the error body
		smContext.Log.Errorf("SMPolicyUpdateNotify err: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	// TODO: Response data type -
	// [200 OK] UeCampingRep
	// [200 OK] array(PartialSuccessReport)
	// [400 Bad Request] ErrorReport
	if err := smContext.ApplyPccRules(decision); err != nil {
		smContext.Log.Errorf("apply sm policy decision error: %+v", err)
		// TODO: Fill the error body
		c.Status(http.StatusBadRequest)
		return
	}

	smContext.SendUpPathChgNotification("EARLY", SendUpPathChgEventExposureNotification)

	ActivateUPFSession(smContext, nil)

	smContext.SendUpPathChgNotification("LATE", SendUpPathChgEventExposureNotification)

	smContext.PostRemoveDataPath()

	c.Status(http.StatusNoContent)
}

func SendUpPathChgEventExposureNotification(
	uri string, notification *models.NsmfEventExposureNotification,
) {
	configuration := EventExposure.NewConfiguration()
	client := EventExposure.NewAPIClient(configuration)
	request := &EventExposure.CreateIndividualSubcriptionMyNotificationPostRequest{
		NsmfEventExposureNotification: notification,
	}
	_, err := client.
		SubscriptionsCollectionApi.
		CreateIndividualSubcriptionMyNotificationPost(context.Background(), uri, request)

	switch err := err.(type) {
	case openapi.GenericOpenAPIError:
		logger.PduSessLog.Warnf("SMF Event Exposure Notification Error[%s]", err.Error())
	case error:
		logger.PduSessLog.Warnf("SMF Event Exposure Notification Failed[%s]", err.Error())
	case nil:
		logger.PduSessLog.Tracef("SMF Event Exposure Notification Success")
	default:
		logger.PduSessLog.Warnf("SMF Event Exposure Notification Unknown Error: %+v", err)
	}
}

func (p *Processor) HandleDNSContextNotify(
	c *gin.Context,
	request models.DnsContextNotification, dnsContextId string,
) {
	logger.PduSessLog.Infoln("In HandleDNSContextNotify")

	c.Status(http.StatusNoContent)

	logger.ConsumerLog.Infof("DNS Context Notify Request Data: %+v", request.EventreportList)

	if request.EventreportList[0].DnsMsgId == "" {
		logger.ConsumerLog.Infof("Only Report not Buffered DNS Message, no need to update DNS Context")
		return
	}

	if request.EventreportList[0].DnsRspReport != nil {
		go p.DecisionMultipleDNAI(context.Background(), &request.EventreportList[0], &dnsContextId)
	} else {
		go p.Consumer().SendDNSContextUpdate(context.Background(), &request.EventreportList[0], &dnsContextId)
	}
}

func (p *Processor) DecisionMultipleDNAI(ctx context.Context, eventReport *models.DnsContextEventReport, dnsContextId *string) error {

	easIpAddresses := eventReport.DnsRspReport.EasIpv4Addresses

	switch factory.SmfConfig.Configuration.Experiment.Type {
	case "Random":
		logger.ProcessorLog.Infof("Experiment Type: %s", factory.SmfConfig.Configuration.Experiment.Type)
		index := rand.Intn(3)
		// index := p.randSlice[p.randIndex]
		// p.randIndex = (p.randIndex + 1) % p.randMax
		logger.ProcessorLog.Infof("Random index: %d", index)
		logger.ProcessorLog.Infof("EAS IP Addresses: %+v", easIpAddresses)
		targetIp := easIpAddresses[index]
		logger.ProcessorLog.Infof("Selected EAS IP Address: %s", targetIp)

		// // test
		// nadafAnalytics, err := p.Consumer().GetNwdafAnalytics()
		// if err != nil {
		// 	logger.ProcessorLog.Errorf("GetNwdafAnalytics error: %+v", err)
		// 	return err
		// }
		// edgeResource, err := p.Consumer().GetEdgeResouceInfo(ctx, "http://127.0.0.163:8000")
		// if err != nil {
		// 	logger.ProcessorLog.Errorf("GetEdgeResouceInfo error: %+v", err)
		// 	return err
		// }

		// decision, err := p.Decisioner.GetDecision(ctx, &nadafAnalytics.DnPerfInfos[0],
		// 	&edgeResource.EdgeResourceInfos, &factory.SmfConfig.EasDeploymentInfo.Dnais)
		// if err != nil {
		// 	logger.ProcessorLog.Errorf("GetDecision error: %+v", err)
		// }

		// logger.ProcessorLog.Infof("GetDecision: %+v", decision)

		p.Consumer().SendEASDecision(ctx, &targetIp, *eventReport, *dnsContextId)

	case "RoundRobin":
		logger.ProcessorLog.Infof("Experiment Type: %s", factory.SmfConfig.Configuration.Experiment.Type)
		p.roundRobinMu.Lock()
		index := p.roundRobin
		p.roundRobin = (p.roundRobin + 1) % p.roundRobinMax
		p.roundRobinMu.Unlock()
		logger.ProcessorLog.Infof("RoundRobin index: %d", index)
		logger.ProcessorLog.Infof("EAS IP Addresses: %+v", easIpAddresses)
		targetIp := easIpAddresses[index]
		logger.ProcessorLog.Infof("Selected EAS IP Address: %s", targetIp)
		p.Consumer().SendEASDecision(ctx, &targetIp, *eventReport, *dnsContextId)
	case "ShortestPath":

		logger.ProcessorLog.Infof("Experiment Type: %s", factory.SmfConfig.Configuration.Experiment.Type)
		targetIp := easIpAddresses[0]
		logger.ProcessorLog.Infof("Selected EAS IP Address: %s", targetIp)
		p.Consumer().SendEASDecision(ctx, &targetIp, *eventReport, *dnsContextId)

	case "SmallestLatency":

		logger.ProcessorLog.Infof("Experiment Type: %s", factory.SmfConfig.Configuration.Experiment.Type)
		nadafAnalytics, err := p.Consumer().GetNwdafAnalytics()
		if err != nil {
			logger.ProcessorLog.Errorf("GetNwdafAnalytics error: %+v", err)
			return err
		}
		dnPerf := nadafAnalytics.DnPerfInfos[0]
		if len(dnPerf.DnPerf) == 0 {
			logger.ProcessorLog.Warnf("No DNAI Performance Data, use RoundRobin as default")
			p.roundRobinMu.Lock()
			p.roundRobin = 0
			p.roundRobinMu.Unlock()
		}

		if p.roundRobin < 3 {
			p.roundRobinMu.Lock()
			index := p.roundRobin
			p.roundRobin++
			p.roundRobinMu.Unlock()
			targetIp := easIpAddresses[index]
			logger.ProcessorLog.Infof("Selected EAS IP Address: %s", targetIp)
			p.Consumer().SendEASDecision(ctx, &targetIp, *eventReport, *dnsContextId)
			return nil
		}

		targetEdgeId := ""
		minDelay := int32(1<<31 - 1)
		for _, edgePerf := range dnPerf.DnPerf {
			if minDelay > edgePerf.PerfData.AvePacketDelay {
				minDelay = edgePerf.PerfData.AvePacketDelay
				targetEdgeId = edgePerf.Dnai
			}
		}
		logger.ProcessorLog.Infof("Selected Edge ID: %s", targetEdgeId)
		index := func(id string) int {
			switch id {
			case "edge1":
				return 2
			case "edge2":
				return 1
			case "edge3":
				return 0
			default:
				return -1
			}
		}(targetEdgeId)

		targetIp := easIpAddresses[index]
		logger.ProcessorLog.Infof("Selected EAS IP Address: %s", targetIp)
		p.Consumer().SendEASDecision(ctx, &targetIp, *eventReport, *dnsContextId)

	case "LLM":
		// test
		nadafAnalytics, err := p.Consumer().GetNwdafAnalytics()
		if err != nil {
			logger.ProcessorLog.Errorf("GetNwdafAnalytics error: %+v", err)
			return err
		}
		edgeResource, err := p.Consumer().GetEdgeResouceInfo(ctx, "http://127.0.0.163:8000")
		if err != nil {
			logger.ProcessorLog.Errorf("GetEdgeResouceInfo error: %+v", err)
			return err
		}

		decisionData, err := p.Decisioner.GetDecision(ctx, &nadafAnalytics.DnPerfInfos[0],
			&edgeResource.EdgeResourceInfos, &factory.SmfConfig.EasDeploymentInfo.Dnais)
		if err != nil {
			logger.ProcessorLog.Errorf("GetDecision error: %+v", err)
		}

		logger.ProcessorLog.Infof("GetDecision: %+v", decisionData)

		decision, ok := decisionData.(smf_model.APIResponse)
		if !ok {
			logger.ProcessorLog.Errorf("GetDecision type assertion failed")
			return fmt.Errorf("GetDecision type assertion failed")
		}
		logger.ProcessorLog.Infof("Selected EAS IP Address: %s", decision.Decision.TargetIPv4)

		p.Consumer().SendEASDecision(ctx, &decision.Decision.TargetIPv4, *eventReport, *dnsContextId)
	}

	// TODO: implement update DNAI decision logic to EASDF

	return nil
}
