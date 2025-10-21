package models

import "time"

type EdcfEdgeResourceDataResponse struct {
	EdgeResourceInfos []EdgeMetric `json:"edgeResourceInfos,omitempty"`
}

type EdgeMetric struct {
	EdgeId    string        `json:"edgeId,omitempty"`
	Period    int64         `json:"period,omitempty"` // in seconds
	Timestamp time.Time     `json:"timestamp,omitempty"`
	CPU       CPUMetric     `json:"cpu,omitempty"`
	Memory    MemoryMetric  `json:"memory,omitempty"`
	Network   NetworkMetric `json:"network,omitempty"`
	Disk      DiskMetric    `json:"disk,omitempty"`
}

type CPUMetric struct {
	Mean float64 `json:"mean,omitempty"`
	P90  float64 `json:"p90,omitempty"`
	Max  float64 `json:"max,omitempty"`
}

type MemoryMetric struct {
	Mean float64 `json:"mean"`
	P90  float64 `json:"p90"`
	Max  float64 `json:"max"`
}

type NetworkMetric struct {
	EgressMbpsMean  float64 `json:"egressMbpsMean,omitempty"`
	EgressMbpsP90   float64 `json:"egressMbpsP90,omitempty"`
	EgressMbpsMax   float64 `json:"egressMbpsMax,omitempty"`
	IngressMbpsMean float64 `json:"ingressMbpsMean,omitempty"`
	IngressMbpsP90  float64 `json:"ingressMbpsP90,omitempty"`
	IngressMbpsMax  float64 `json:"ingressMbpsMax,omitempty"`
	TrafficPattern  string  `json:"trafficPattern,omitempty"`
}

type DiskMetric struct {
	ReadBpsMean  float64 `json:"readBpsMean,omitempty"`
	WriteBpsMean float64 `json:"writeBpsMean,omitempty"`
}
