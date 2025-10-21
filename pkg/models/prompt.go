package models

import (
	"time"

	"github.com/free5gc/openapi/models"
)

// LLMDecisionPrompt 是送進 LLM 的輸入格式
type LLMDecisionPrompt struct {
	Context    string              `json:"context"`            // 給 LLM 的任務描述
	UePerfInfo []models.DnPerfInfo `json:"uePerformance"`      // UE Performance 資料
	EdgeInfos  []EdgeMetric        `json:"edgeResources"`      // Edge Resource 資料
	Question   string              `json:"question,omitempty"` // 你想問 LLM 的決策問題，例如 "哪個 Edge 最適合?"
}

// LLMDecision 是 LLM 回傳給 EASDF/NWDAF 的最終決策
type LLMDecision struct {
	Version      string            `json:"version,omitempty"`      // 決策格式版本
	DecisionID   string            `json:"decisionId,omitempty"`   // 追蹤用
	Timestamp    time.Time         `json:"timestamp"`              // 產生時間（UTC）
	Type         DecisionType      `json:"type"`                   // assign/defer/reject
	Target       *DecisionTarget   `json:"target,omitempty"`       // 被指派的 Edge（type=assign 時必填）
	Ranking      []RankedCandidate `json:"ranking,omitempty"`      // 候選清單（排序）
	Constraints  *DecisionPolicy   `json:"constraints,omitempty"`  // 這次判斷參考/滿足的需求
	Validity     *ValidityWindow   `json:"validity,omitempty"`     // 決策有效期間
	SpatialScope *SpatialScope     `json:"spatialScope,omitempty"` // 適用的區域（可選）
	Confidence   float64           `json:"confidence,omitempty"`   // 0~1 對本決策的置信度
	Rationale    []Reason          `json:"rationale,omitempty"`    // 可解釋理由（多條）
	Fallback     *FallbackPlan     `json:"fallback,omitempty"`     // 失敗/退避策略
	Notes        []string          `json:"notes,omitempty"`        // 其他備註
}

// ---- 基本欄位型別 ----

type DecisionType string

const (
	DecisionAssign DecisionType = "assign" // 指派某一個 Edge
	DecisionDefer  DecisionType = "defer"  // 暫緩（需要更多資料或觀察）
	DecisionReject DecisionType = "reject" // 拒絕（沒有任何合適的 Edge）
)

type DecisionTarget struct {
	EdgeID string  `json:"edgeId"`           // 最終選擇
	Score  float64 `json:"score,omitempty"`  // 合成分數（0~1；非硬性）
	Reason string  `json:"reason,omitempty"` // 一句話摘要
}

type RankedCandidate struct {
	EdgeID       string   `json:"edgeId"`
	Score        float64  `json:"score,omitempty"`  // 合成分數（與 Ranking 對齊）
	CPUHint      *float64 `json:"cpuP90,omitempty"` // 可回傳部分關鍵特徵，便於驗證
	MemHint      *float64 `json:"memP90,omitempty"`
	EgressHint   *float64 `json:"egressMbpsP90,omitempty"`
	HeadroomHint *float64 `json:"headroom,omitempty"`
	Reasons      []Reason `json:"reasons,omitempty"` // 該候選的局部理由
}

type DecisionPolicy struct {
	MinHeadroom       *float64 `json:"minHeadroom,omitempty"`       // 例：0.15
	MaxLatencyMsP90   *int32   `json:"maxLatencyMsP90,omitempty"`   // 例：50
	MaxPktLossPpmP90  *int32   `json:"maxPktLossPpmP90,omitempty"`  // 例：3000 (=0.3%)
	MinThroughputMbps *int32   `json:"minThroughputMbps,omitempty"` // 例：100
	PreferStable      bool     `json:"preferStable,omitempty"`      // 偏好穩定性
}

type ValidityWindow struct {
	Start time.Time `json:"start"` // 決策開始生效
	End   time.Time `json:"end"`   // 建議在此時間後重評
}

type SpatialScope struct {
	DNAI  string   `json:"dnai,omitempty"`    // 與 DNPerf 對齊
	TAI   []string `json:"taiList,omitempty"` // 可選：TAI 列表
	Cells []string `json:"cellIds,omitempty"` // 可選：小區
}

type Reason struct {
	Code    ReasonCode `json:"code"`              // 結構化代碼（便於機器處理）
	Message string     `json:"message,omitempty"` // 可讀文字
}

type ReasonCode string

const (
	RCPUHeadroomGood    ReasonCode = "CPU_HEADROOM_GOOD"
	RMemHeadroomGood    ReasonCode = "MEM_HEADROOM_GOOD"
	REgressCapacityGood ReasonCode = "EGRESS_CAPACITY_GOOD"
	RStabilityGood      ReasonCode = "STABILITY_GOOD"
	RTrendDecreasing    ReasonCode = "TREND_DECREASING"
	RQoSBetterLatency   ReasonCode = "QOS_BETTER_LATENCY"
	RQoSBetterLoss      ReasonCode = "QOS_BETTER_LOSS"
	ROverloaded         ReasonCode = "OVERLOADED"
	RInsufficientData   ReasonCode = "INSUFFICIENT_DATA"
	RPolicyViolation    ReasonCode = "POLICY_VIOLATION"
)

type FallbackPlan struct {
	TryNext         []string     `json:"tryNext,omitempty"`            // 失敗時依序嘗試的 edgeId
	ReevaluateAfter *int32       `json:"reevaluateAfterSec,omitempty"` // 幾秒後重評
	Trigger         []ReasonCode `json:"trigger,omitempty"`            // 何種狀況觸發退避
}
