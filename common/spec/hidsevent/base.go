package hidsevent

import "encoding/json"

type HIDSEvent string
type HIDSTimestampType string

type HIDSSoftwareType string

var (
	// process monitoring
	HIDSEvent_Proccessings     HIDSEvent = "proccessings"
	HIDSEvent_ProccessingEvent HIDSEvent = "processing-event"
	HIDSEvent_ProccessingTouch HIDSEvent = "processing-touch"

	// network connection status
	// Handle with caution :
	//    1. Note that in some cases the server will have a lot of connections from the outside to the inside. If this situation is over-handled, it will cause excessive resource consumption.
	//    2. You can refer to telegrafs use of gopsutil
	//    3. If necessary, copy the code and be sure to avoid the problems caused in (1): You can check if there is paging/or pre-screening method
	// needs to handle the following connections:
	// 1. The local listening port (the port whose Status is LISTEN) is reported. ) (Key points)
	// 2. Ports for external connections (it is estimated that there will not be too many, this must be processed) (Key points)
	// 3. Try to avoid HA/NGINX The impact of excessive connections on monitoring
	//
	HIDSEvent_Connections     HIDSEvent = "connections"
	HIDSEvent_ConnectionTouch HIDSEvent = "connections-touch"
	HIDSEvent_ConnectionEvent HIDSEvent = "connection-event"

	// nginx / apache monitors
	HIDSEvent_NginxFound   HIDSEvent = "nginx-found"
	HIDSEvent_NginxMissed  HIDSEvent = "nginx-missed"
	HIDSEvent_ApacheFound  HIDSEvent = "apache-found"
	HIDSEvent_ApacheMissed HIDSEvent = "apache-missed"

	// ssh audit analysis
	// 1. obtain SSH accurate version information
	// 2. Configuration file, public key and private key monitoring
	// 3. Configuration file key options:
	//    1. Whether to allow password login
	//    2. Whether to allow empty passwords
	//    3. Key login
	HIDSEvent_SSHAudit HIDSEvent = "ssh-audit"

	// file changes
	// temporary default monitoring /etc /bin /usr/bin ~/File contents under .ssh
	HIDSEvent_FileChanged HIDSEvent = `file-changed`

	// monitors the webshell
	HIDSEvent_WebShell HIDSEvent = "webshell"

	// nodes are scanned (NIDS function, optional)
	HIDSEvent_Scanned HIDSEvent = "scanned"

	// key configuration file
	HIDSEvent_ConfigFile HIDSEvent = "config-file"

	// vulnerability information
	HIDSEvent_VulnInfo HIDSEvent = "vuln-info"

	// Dangerous file samples
	HIDSEvent_DangerousFileSample HIDSEvent = "dangerous-file-sample"

	// attack behavior
	HIDSEvent_Attack HIDSEvent = "attack"

	HIDSEvent_ReverseShell HIDSEvent = "reverse-shell"

	//Request configuration
	HIDSEvent_RequestConfig HIDSEvent = "request_config"

	//reporting host user information
	HIDSEvent_ReportHostUser HIDSEvent = "report_host_user"
	//reports all user information that successfully logged in
	HIDSEvent_ReportAllUsrLoginOK HIDSEvent = "all_user_login_ok"
	//Report all user information that failed to log in
	HIDSEvent_ReportAllUsrLoginFail HIDSEvent = "all_user_login_fail"
	//reports all user information files that failed to log in Too large
	HIDSEvent_ReportAllUsrLoginFailFileTooLarge HIDSEvent = "all_user_login_fail_file_too_large"
	//user account brute force attack
	HIDSEvent_UserLoginAttempt HIDSEvent = "user_login_attempt"
	//software information reporting
	HIDSEvent_ReportSoftwareVersion HIDSEvent = "report_software_version"
	//enable startup software information
	HIDSEvent_BootSoftware HIDSEvent = "boot_software"
	//scheduled task
	HIDSEvent_Crontab HIDSEvent = "crontab"
)

var (
	HIDSEvent_Notify_Config HIDSEvent = "notify_config"
)

var (
	HIDSTimestampType_Last_Check_Login_Fail HIDSTimestampType = "last_check_login_fail"
)

var (
	HIDSSoftwareType_APT HIDSSoftwareType = "apt"
	HIDSSoftwareType_YUM HIDSSoftwareType = "yum"
)

type HIDSMessage struct {
	Type    HIDSEvent       `json:"event"`
	Content json.RawMessage `json:"content"`
}
