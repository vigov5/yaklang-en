package spec

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yaklang/yaklang/common/fp"
	"github.com/yaklang/yaklang/common/utils"
	"net"
	"net/http"
)

type ScanPortTask struct {
	// Scan target
	Hosts string `json:"host"`
	Ports string `json:"port"`
}

type ScanResultType string

const (
	// Only port open information
	ScanResult_PortState ScanResultType = "port_state"

	// Fp.MatcherResult contains fingerprint information
	ScanResult_Fingerprint ScanResultType = "fingerprint"

	// *yakit.Report overall report
	ScanResult_Report ScanResultType = "report"

	// HttpFlow asset information
	ScanResult_HTTPFlow ScanResultType = "http-flow"

	// Vulnerability information, weak passwords and the like should also be included in this
	ScanResult_Vuln ScanResultType = "vuln"

	// Discover domain name assets and other information
	ScanResult_Domain ScanResultType = "domain"
)

type ScanResult struct {
	Type    ScanResultType  `json:"type"`
	Content json.RawMessage `json:"content"`

	// If these three fields are present, it means that it is a distributed task, and this content needs to be processed additionally
	TaskId    string `json:"task_id"`
	RuntimeId string `json:"runtime_id"`
	SubTaskId string `json:"sub_task_id"`
}

type PortStateType string

const (
	PortStateType_Unknown PortStateType = "unknown"
	PortStateType_Open    PortStateType = "open"
	PortStateType_Closed  PortStateType = "closed"
)

type PortFingerprint struct {
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	Proto       fp.TransportProto `json:"proto"`
	State       PortStateType     `json:"state"`
	CPEs        []string          `json:"cpes"`
	Banner      string            `json:"banner"`
	ServiceName string            `json:"service_name"`
}

func NewHTTPFlowScanResult(isHttps bool, req *http.Request, rsp *http.Response) (*ScanResult, error) {
	return nil, utils.Error("not implemented")
}

func NewScanFingerprintResult(m *fp.MatchResult) (*ScanResult, error) {
	if m.Fingerprint == nil {
		return nil, errors.Errorf("fetch fingerprint to %v failed: %v", m.Target, m.Reason)
	}

	var state PortStateType
	switch m.State {
	case fp.OPEN:
		state = PortStateType_Open
	case fp.CLOSED:
		state = PortStateType_Closed
	case fp.UNKNOWN:
		state = PortStateType_Unknown
	default:
		state = PortStateType_Unknown
	}

	f := &PortFingerprint{
		Host:        m.Target,
		Port:        m.Port,
		Proto:       m.GetProto(),
		State:       state,
		CPEs:        m.GetCPEs(),
		Banner:      m.GetBanner(),
		ServiceName: m.GetServiceName(),
	}

	raw, err := json.Marshal(f)
	if err != nil {
		return nil, errors.Errorf("marshal port fingerprint failed: %v", err)
	}

	return &ScanResult{
		Type:    ScanResult_Fingerprint,
		Content: raw,
	}, nil
}

type PortState struct {
	Host  string            `json:"host"`
	Port  int               `json:"port"`
	Proto fp.TransportProto `json:"proto"`
	State PortStateType     `json:"state"`
}

func NewScanTCPOpenPortResult(ip net.IP, port int, state PortStateType) (*ScanResult, error) {
	raw, err := json.Marshal(&PortState{
		Host:  ip.String(),
		Port:  port,
		Proto: fp.TCP,
		State: state,
	})
	if err != nil {
		return nil, errors.Errorf("marshal port state failed: %v", err)
	}

	return &ScanResult{
		Type:    ScanResult_PortState,
		Content: raw,
	}, nil
}

func (p *PortState) String() string {
	var prefix string
	switch p.State {
	case PortStateType_Closed:
		prefix = "CLOSED"
	case PortStateType_Open:
		prefix = "  OPEN"
	case PortStateType_Unknown:
		prefix = "UNKNOW"
	default:
		prefix = "UNKNOW"
	}
	return fmt.Sprintf("%v: [%v] %15s:%v", p.Proto, prefix, p.Host, p.Port)
}
