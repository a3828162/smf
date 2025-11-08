package consumer

import (
	"crypto/tls"
	"net/http"

	"github.com/free5gc/openapi/amf/Communication"
	"github.com/free5gc/openapi/chf/ConvergedCharging"
	"github.com/free5gc/openapi/easdf/DNSContext"
	"github.com/free5gc/openapi/nrf/NFDiscovery"
	"github.com/free5gc/openapi/nrf/NFManagement"
	Nnwdaf_AnalyticsInfo "github.com/free5gc/openapi/nwdaf/AnalyticsInfo"
	"github.com/free5gc/openapi/pcf/SMPolicyControl"
	"github.com/free5gc/openapi/smf/PDUSession"
	"github.com/free5gc/openapi/udm/SubscriberDataManagement"
	"github.com/free5gc/openapi/udm/UEContextManagement"
	"github.com/free5gc/smf/pkg/app"
)

type Consumer struct {
	app.App

	// consumer services
	*nsmfService
	*namfService
	*nchfService
	*npcfService
	*nudmService
	*nnrfService
	*neasdfService
	*nnwdafService
}

func NewConsumer(smf app.App) (*Consumer, error) {
	c := &Consumer{
		App: smf,
	}

	c.nsmfService = &nsmfService{
		consumer:          c,
		PDUSessionClients: make(map[string]*PDUSession.APIClient),
	}

	c.namfService = &namfService{
		consumer:             c,
		CommunicationClients: make(map[string]*Communication.APIClient),
	}

	c.nchfService = &nchfService{
		consumer:                 c,
		ConvergedChargingClients: make(map[string]*ConvergedCharging.APIClient),
	}

	c.nudmService = &nudmService{
		consumer:                        c,
		SubscriberDataManagementClients: make(map[string]*SubscriberDataManagement.APIClient),
		UEContextManagementClients:      make(map[string]*UEContextManagement.APIClient),
	}

	c.nnrfService = &nnrfService{
		consumer:            c,
		NFManagementClients: make(map[string]*NFManagement.APIClient),
		NFDiscoveryClients:  make(map[string]*NFDiscovery.APIClient),
	}

	c.npcfService = &npcfService{
		consumer:               c,
		SMPolicyControlClients: make(map[string]*SMPolicyControl.APIClient),
	}

	c.neasdfService = &neasdfService{
		consumer:            c,
		DNSContextClients:   make(map[string]*DNSContext.APIClient),
		UEIpDNSContextIdMap: make(map[string]string),
	}

	c.nnwdafService = &nnwdafService{
		consumer:             c,
		analyticsInfoClients: make(map[string]*Nnwdaf_AnalyticsInfo.APIClient),
		edgeResouceInfoClients: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}

	return c, nil
}
