package consumer

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/free5gc/openapi/easdf/DNSContext"
	Neasdf_DNSContext "github.com/free5gc/openapi/easdf/DNSContext"
	"github.com/free5gc/openapi/models"
	smf_context "github.com/free5gc/smf/internal/context"
	"github.com/free5gc/smf/pkg/factory"

	"github.com/free5gc/smf/internal/logger"
)

type neasdfService struct {
	consumer *Consumer

	DNSContextMu sync.RWMutex

	DNSContextClients map[string]*DNSContext.APIClient

	UEIpDNSContextIdMap map[string]string
}

func (s *neasdfService) GetDNSContextClient(uri string) *DNSContext.APIClient {
	if uri == "" {
		return nil
	}
	s.DNSContextMu.RLock()
	client, ok := s.DNSContextClients[uri]
	if ok {
		s.DNSContextMu.RUnlock()
		return client
	}

	configuration := DNSContext.NewConfiguration()
	configuration.SetBasePath(uri)
	client = DNSContext.NewAPIClient(configuration)

	s.DNSContextMu.RUnlock()
	s.DNSContextMu.Lock()
	defer s.DNSContextMu.Unlock()
	s.DNSContextClients[uri] = client
	return client
}

// BuildDnsContextCreateData builds DnsContextCreateData from EasDeploymentInfo configuration
func (s *neasdfService) BuildDnsContextCreateData(
	smContext *smf_context.SMContext,
	easDeploymentInfo *factory.EasDeploymentInfo,
) (*models.DnsContextCreateData, error) {
	if easDeploymentInfo == nil {
		return nil, fmt.Errorf("easDeploymentInfo is nil")
	}

	notifyURI := fmt.Sprintf("%s://%s:%d%s/dns-contexts/notify",
		smf_context.GetSelf().URIScheme,
		smf_context.GetSelf().RegisterIPv4,
		smf_context.GetSelf().SBIPort,
		factory.SmfCallbackUriPrefix)

	dnsRules := make(map[string]models.DnsRule)
	precedence := int32(5)

	// Convert BaselineDnsPatterns to DnsRules
	for _, pattern := range easDeploymentInfo.BaselineDnsPatterns {
		dnsRule := models.DnsRule{
			DnsRuleId:       pattern.Id,
			Precedence:      precedence,
			DnsQueryMdtList: make(map[string]models.DnsQueryMdt),
			DnsRspMdtList:   make(map[string]models.DnsRspMdt),
			ActionList:      make(map[string]models.Action),
		}

		// Handle DNS Query Pattern
		if pattern.Type == "QUERY" && pattern.DnsQueryMdt != nil {
			dnsQueryMdt := models.DnsQueryMdt{
				FqdnPatternList: []models.FqdnPatternMatchingRule{},
			}

			for _, fqdnPattern := range pattern.DnsQueryMdt.FqdnPatterns {
				matchingRule := models.FqdnPatternMatchingRule{}
				if fqdnPattern.Regx != "" {
					matchingRule.Regex = fqdnPattern.Regx
				}
				dnsQueryMdt.FqdnPatternList = append(dnsQueryMdt.FqdnPatternList, matchingRule)
			}

			dnsRule.DnsQueryMdtList[pattern.DnsQueryMdt.MdtId] = dnsQueryMdt
		}

		// Handle DNS Response Pattern
		if pattern.Type == "RESPONSE" && pattern.DnsRspMdt != nil {
			dnsRspMdt := models.DnsRspMdt{
				FqdnPatternList: []models.FqdnPatternMatchingRule{},
			}

			// Add FQDN patterns
			for _, fqdnPattern := range pattern.DnsRspMdt.FqdnPatterns {
				matchingRule := models.FqdnPatternMatchingRule{}
				if fqdnPattern.Regx != "" {
					matchingRule.Regex = fqdnPattern.Regx
				}
				dnsRspMdt.FqdnPatternList = append(dnsRspMdt.FqdnPatternList, matchingRule)
			}

			Ipv4AddressRanges := []models.Ipv4AddressRange{}
			for _, ipv4Range := range pattern.DnsRspMdt.EasIpv4AddrRanges {
				ipv4AddrRange := models.Ipv4AddressRange{}
				if ipv4Range.Start != "" {
					ipv4AddrRange.Start = ipv4Range.Start
				}
				if ipv4Range.End != "" {
					ipv4AddrRange.End = ipv4Range.End
				}
				Ipv4AddressRanges = append(Ipv4AddressRanges, ipv4AddrRange)
			}

			dnsRspMdt.EasIpv4AddrRanges = Ipv4AddressRanges
			// Note: IP address ranges handling depends on EASDF openapi models
			// For now, we focus on FQDN patterns as per the configuration

			dnsRule.DnsRspMdtList[pattern.DnsRspMdt.MdtId] = dnsRspMdt
		}

		// Handle DNS Action

		if pattern.DnsAction != nil {
			for _, a := range pattern.DnsAction {
				action := models.Action{}

				switch a.ApplyAction {
				case "REPORT":
					action.ApplyAction = models.ApplyAction_REPORT
					action.ReportingOnceInd = a.ReportingOnceInd
				case "FORWARD":
					action.ApplyAction = models.ApplyAction_FORWARD
					// Add forwarding parameters
					if a.DnsServerAddr != "" {
						action.FwdParas = &models.ForwardingParameters{
							DnsServerAddressInfo: &models.DnsServerAddressInfo{
								DnsServerAddressList: []models.IpAddr{
									{
										Ipv4Addr: a.DnsServerAddr,
									},
								},
							},
						}
					} else if easDeploymentInfo.LocalDNSServer.IPv4 != "" {
						// Use local DNS server if no specific server is provided
						action.FwdParas = &models.ForwardingParameters{
							DnsServerAddressInfo: &models.DnsServerAddressInfo{
								DnsServerAddressList: []models.IpAddr{
									{
										Ipv4Addr: easDeploymentInfo.LocalDNSServer.IPv4,
									},
								},
							},
						}
					}
				case "BUFFER":
					action.ApplyAction = models.ApplyAction_BUFFER
				}

				dnsRule.ActionList[a.ActionId] = action
			}

		}

		dnsRules[pattern.Id] = dnsRule
	}

	dnsContextCreateData := &models.DnsContextCreateData{
		UeIpv4Addr: smContext.PDUAddress.String(),
		Dnn:        smContext.Dnn,
		SNssai: &models.Snssai{
			Sst: smContext.SNssai.Sst,
			Sd:  smContext.SNssai.Sd,
		},
		DnsRules:  dnsRules,
		NotifyUri: notifyURI,
	}

	return dnsContextCreateData, nil
}

func (s *neasdfService) SendDNSContextCreate(ctx context.Context, smContext *smf_context.SMContext) (
	string, error,
) {
	var dnsContextCreateData *models.DnsContextCreateData
	var err error

	// Try to use EasDeploymentInfo from configuration
	if factory.SmfConfig.EasDeploymentInfo != nil && len(factory.SmfConfig.EasDeploymentInfo.BaselineDnsPatterns) > 0 {
		logger.ConsumerLog.Infoln("Using EasDeploymentInfo from configuration")
		dnsContextCreateData, err = s.BuildDnsContextCreateData(smContext, factory.SmfConfig.EasDeploymentInfo)
		if err != nil {
			logger.ConsumerLog.Warnf("Failed to build DNS context from config: %v, falling back to hardcoded rules", err)
			dnsContextCreateData = nil
		}
	}

	// Fallback to hardcoded rules if no configuration is available
	// if dnsContextCreateData == nil {
	// 	logger.ConsumerLog.Infoln("Using hardcoded DNS rules")
	// 	notifyURI := fmt.Sprintf("%s://%s:%d%s/dns-contexts/notify",
	// 		smf_context.GetSelf().URIScheme,
	// 		smf_context.GetSelf().RegisterIPv4,
	// 		smf_context.GetSelf().SBIPort,
	// 		factory.SmfCallbackUriPrefix)

	// 	dnsContextCreateData = &models.DnsContextCreateData{
	// 		UeIpv4Addr: smContext.PDUAddress.String(),
	// 		Dnn:        smContext.Dnn,
	// 		SNssai: &models.Snssai{
	// 			Sst: smContext.SNssai.Sst,
	// 			Sd:  smContext.SNssai.Sd,
	// 		},
	// 		DnsRules: map[string]models.DnsRule{
	// 			"1": {
	// 				DnsRuleId:  "1",
	// 				Precedence: 1,
	// 				DnsQueryMdtList: map[string]models.DnsQueryMdt{
	// 					"VR_Stream": {
	// 						FqdnPatternList: []models.FqdnPatternMatchingRule{
	// 							{
	// 								Regex: "hhhh.com.",
	// 							},
	// 						},
	// 					},
	// 				},
	// 				DnsRspMdtList: map[string]models.DnsRspMdt{},
	// 				ActionList: map[string]models.Action{
	// 					"1": {
	// 						ApplyAction: models.ApplyAction_FORWARD,
	// 						FwdParas: &models.ForwardingParameters{
	// 							DnsServerAddressInfo: &models.DnsServerAddressInfo{
	// 								DnsServerAddressList: []models.IpAddr{
	// 									{
	// 										Ipv4Addr: "192.168.56.50",
	// 									},
	// 								},
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 		NotifyUri: notifyURI,
	// 	}
	// }

	var dnsReq Neasdf_DNSContext.CreateDnsContextRequest
	dnsReq.DnsContextCreateData = dnsContextCreateData

	logger.ConsumerLog.Infoln("Send DNSContext Create Request to EASDF")
	logger.ConsumerLog.Infof("Request Data: %+v", dnsReq.DnsContextCreateData)

	// Get EASDF client (use configured IP or default)
	easdfURI := "http://127.0.0.56:8000"

	client := s.GetDNSContextClient(easdfURI)
	if client == nil {
		err := fmt.Errorf("DNSContext Client is nil")
		logger.ConsumerLog.Errorf("Create DNSContext Error: %s", err)
		return "", err
	}

	res, err := client.DNSContextsCollectionApi.CreateDnsContext(ctx, &dnsReq)
	if err != nil {
		logger.ConsumerLog.Errorf("Create DNSContext Error: %s", err)
		return "", err
	}

	lastIndex := strings.LastIndex(res.Location, "/")

	logger.ConsumerLog.Infoln("Response Created Data: ", res.DnsContextCreatedData)
	logger.ConsumerLog.Infoln("Response Location: ", res.Location)
	logger.ConsumerLog.Infoln("Create DNSCtx Success")

	s.UEIpDNSContextIdMap[smContext.PDUAddress.String()] = res.Location[lastIndex+1:]

	return res.Location[lastIndex+1:], nil
}
