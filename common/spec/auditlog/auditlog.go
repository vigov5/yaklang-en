package auditlog

import (
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Level string

const (
	INFO    Level = "info"
	DEBUG   Level = "debug"
	TRACE         = DEBUG
	WARN    Level = "info"
	WARNING       = WARN
	ERROR   Level = "error"
	FATAL   Level = "fatal"
	PANIC   Level = "panic"
)

type EventSeverity int

const (
	LogNormal     EventSeverity = 1
	LogMiddleLow  EventSeverity = 2
	LogMiddle     EventSeverity = 3
	LogMiddleHigh EventSeverity = 4
	LogHigh       EventSeverity = 5
)

func (e *EventSeverity) String() string {
	switch *e {
	case LogNormal:
		return "info"
	case LogMiddleLow:
		return "low"
	case LogMiddle:
		return "middle"
	case LogMiddleHigh:
		return "high"
	case LogHigh:
		return "alarm"
	}
	return ""
}

type AuditLog struct {
	/*
	   The attributes of the log itself
	*/
	// Logagent Query log specific time
	QueryTimestamp int64 `json:"query_timestamp" `

	LogRecordID uint `json:"record_id"`
	//The unique ID of the URL requested
	RequestId string `json:"request_id"`

	// Log record specific timestamp
	Timestamp int64 `json:"log_timestamp" gorm:"index"`

	// level
	Level EventSeverity `json:"event_severity"`

	/*
		The content of the log record (who - when - what is used/(it occurs in that interface: url / source ) - What was done Event - What additional data is generated/Impact)
	*/

	// The operating user, which may be the user name or the system name
	OperationUser string `json:"operation_user" gorm:"index"` // Who
	Organization  string // If the persons organization
	VpnId         string `json:"vpn_account" gorm:"index"` // If the log generated through VPN operation

	// occur in?
	UrlPath  string `json:"url_path" gorm:"index"` // Log generated for URL
	Source   string `json:"log_type" gorm:"index"` // In which system does the log occur?
	SpotInfo string `json:"spot_info"`             // Event information - Which module? Which file? Which function?

	// Network and protocol related content
	// Source and destination network address of the operation
	DstIP   string `json:"dst_ip"`
	DstPort int    `json:"dst_port"`
	SrcIP   string `json:"src_ip"`
	SrcPort int    `json:"src_port"`

	// http
	HttpMethod          string         `json:"http_method"`
	HttpResponseCode    int            `json:"http_response_code"`
	HttpContentType     string         `json:"http_content_type"`
	HttpContentLength   int            `json:"http_content_length"`
	HttpClientUserAgent string         `json:"http_client_user_agent"`
	HttpHost            string         `json:"http_host"`
	HttpRequestBody     postgres.Jsonb `json:"http_request_body" `

	// Log content
	//Content map[string]interface{} `json:"content" gorm:"type: jsonb"`
	Content postgres.Jsonb `json:"content"  `

	// This ExtraData represents the log content extracted from the log content
	// If the content contains ID card, mobile phone number and other information, and is captured or analyzed by regular expressions, the extracted data will be structured and put into ExtraData.
	// or JSON is extracted, it will also be extracted and put into the extra data. Where is
	ExtraData postgres.Jsonb `json:"extra" `
	//Token returned by sso login
	BetaUserToken string `json:"beta_user_token" `
	DeptPath      string `json:"dept_path"` //Department path
	PsnStatus     string `json:"psnStatus"` //Personnel status 1 Incumbent 0 Resigned
}

type RpmsPerson struct {
	WorkCity       string      `json:"workCity"`      //Working city
	PsnStatus      string      `json:"psnStatus"`     //Personnel status 1 Incumbent 0 Resigned
	IdType         string      `json:"idType"`        //ID type
	Org            string      `json:"org"`           //Organization
	Name           string      `json:"name"`          //Employee name
	MobileMd5      string      `json:"mobileMd5"`     //Phone number MD5 before desensitization
	Mobile         string      `json:"mobile"`        //Contact number
	JoinDate       interface{} `json:"joinDate"`      //Joining date
	IdMd5          string      `json:"idMd5"`         //ID number MD5 value before desensitization
	IdNo           string      `json:"idNo"`          //ID number
	Email          string      `json:"email"`         //Corporate email address
	DispatchCorp   string      `json:"dispatchCorp"`  //Dispatch company
	DismissingDate interface{} `json:"dimissionDate"` //Resigning date
	DeptPath       string      `json:"deptPath"`      //Department path
	Dept           string      `json:"dept"`          //Department
	DeptLevel1     string      `json:"dept_level1"`   //First-level department
	DeptLevel2     string      `json:"dept_level2"`   //Secondary department
	DeptLevel3     string      `json:"dept_level3"`   //Third-level department
	Code           string      `json:"code"`          //Employee ID
	PsnClass       string      `json:"psnClass"`      //Employee type
}

type SsoLogin struct {
	UserId        string `json:"user_id"`        // The unique ID of the user
	Email         string `json:"email"`          // Sso login email account account number
	LoginIp       string `json:"login_ip"`       // Login IP information
	TargetSystem  string `json:"target_system"`  // Login target system
	DeviceId      string `json:"device_id"`      // Device ID for mobile phone login
	FingerPrint   string `json:"fingerprint"`    //Device ID for browser login
	LoginCountry  string `json:"login_country"`  // Geographical attribute of login IP
	LoginProvince string `json:"login_province"` //
	LoginCity     string `json:"login_city"`     //

}

type BI struct {
	TracerReportId   string `json:"tracerReportId"`   // Report id
	TracerReportName string `json:"tracerReportName"` //Report name
	DateKey          string `json:"datekey"`          //Time (20200619)
	AreaInfo         string `json:"areaInfo"`         //Region , City
	AreaName         string `json:"areaName"`         //Region name
	ClassInfo        string `json:"classInfo"`        //Category
	ClassName        string `json:"className"`        //category Name
	MmcInfo          string `json:"mmcInfo"`          //Merchant attribution
	MmcName          string `json:"mmcName"`          //Merchant name
	CustomerType     string `json:"customerType"`     //Family restaurant code
	CustomerName     string `json:"customerName"`     //Family/Individual, restaurant
}

type Authentication struct {
	UserId        string `json:"user_id"`         // The unique ID of the user
	TargetUrlPath string `json:"target_url_path"` // Target url accessed
	TargetSystem  string `json:"target_system"`   // The system to which the accessed url belongs
	AccessResult  bool   `json:"access_result"`   // Authentication result, true=can access false=denied access
	RealIp        string `json:"real_ip"`         // The real originating IP address of the access
	ForwardIp     string `json:"forward_ip"`      // IP of forwarding service
}

// DingTalk report message, use
type DingReportMsg struct {
	Date          string `json:"date"`
	Name          string `json:"name"`
	DataNum       int    `json:"data_num"`
	TopDeptNum    int    `json:"top_dept_num"`
	WorkCity      int    `json:"work_city"`
	BottomDeptNum int    `json:"bottom_deptNum"`
	MobileNum     int    `json:"mobile_num"`
	OrgNum        int    `json:"org_num"`
	IdNum         int    `json:"id_num"`
	IsWhiteRole   bool   `json:"is_white_role"`
}

type DingReportMsgList []*DingReportMsg

func (p DingReportMsgList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p DingReportMsgList) Len() int      { return len(p) }
func (p DingReportMsgList) Less(i, j int) bool {
	if p[i].IsWhiteRole && p[j].IsWhiteRole {
		return p[i].DataNum > p[j].DataNum
	}
	if p[i].IsWhiteRole {
		return false
	}
	if p[j].IsWhiteRole {
		return true
	}
	return p[i].DataNum > p[j].DataNum
}

type PairKeyStringValueInt struct {
	Key   string
	Value int
}

type PairKeyStringValueIntList []*PairKeyStringValueInt

func (p PairKeyStringValueIntList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairKeyStringValueIntList) Len() int           { return len(p) }
func (p PairKeyStringValueIntList) Less(i, j int) bool { return p[i].Value > p[j].Value }
