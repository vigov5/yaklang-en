package spec

import (
	"fmt"
	"strings"
	"time"
)

var (
	LocalTimeLocation *time.Location
)

func init() {
	var err error
	LocalTimeLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		//log.Infof("load localtime failed: %s", err)
	}
}

type ServerPushKey string

const (
	// Node ID of the server Constant
	ServerNodeId   = "palm-server"
	LogAgentNodeId = "log-agent"

	// RPC Exchange Must be Direct
	CommonRPCExchange = "palm-rpc"

	// Switch to which the server pushes data
	CommonServerPushExchange   = "palm-push"
	CommonServerPushDefaultKey = "palm.nodebase.notification"

	// Key to which the server pushes data to the switch
	ServerPush_ScriptTask ServerPushKey = "script-task"

	ScanPortExchange        = "palm-scan-port-task"
	ScanFingerprintExchange = ScanPortExchange

	// scan port
	CommonScanPortQueue     = "palm-scan-port"
	CommonScanPortTaskKey   = "palm.stream.task.scan-port"
	CommonScanPortResultKey = "palm.stream.result.scan-port"

	CommonScanFingerprintQueue     = "palm-scan-fingerprint"
	CommonScanFingerprintTaskKey   = "palm.stream.task.scan-fingerprint"
	CommonScanFingerprintResultKey = "palm.stream.result.scan-fingerprint"

	API_RegisterNode   = "register-palm-node"
	API_UnregisterNode = "unregister-palm-node"

	BackendKey_HTTPFlow                           = "http-flow"
	BackendKey_Scanner                            = "scanner"
	BackendKey_ProcessInfo                        = "process"
	BackendKey_ProcessEvent                       = "process-event"
	BackendKey_ConnectionEvent                    = "connection-event"
	BackendKey_NetConnectInfo                     = "netconnect"
	BackendKey_Nginx                              = "nginx"
	BackendKey_Apache                             = "apache"
	BackendKey_FileChangeInfo                     = "filechange"
	BackendKey_SystemMatrix                       = "heartbeat"
	BackendKey_SSHInfo                            = "sshinfo"
	BackendKey_RequestConfig                      = "request_config"
	BackendKey_ReportHostUser                     = "report_host_user"
	BackendKey_ReportAllUserLoginOk               = "report_all_user_login_ok"
	BackendKey_ReportAllUserLoginFail             = "report_all_user_login_fail"
	BackendKey_ReportAllUserLoginFailFileTooLarge = "report_all_user_login_fail_file_too_large"
	BackendKey_Heartbeat                          = BackendKey_SystemMatrix
	BackendKey_UserLoginAttempt                   = "user_login_attempt"
	BackendKey_SoftwareVersion                    = "software_version"
	BackendKey_BootSoftware                       = "boot_software"
	BackendKey_Crontab                            = "crontab"
	BackendKey_ReverseShell                       = "reverse_shell"
	BackendKey_NodeLog                            = "node_log"

	HIDS_API_Sleep = "hids-rpc-sleep"
)

var (
	HIDS_APIs = []string{
		HIDS_API_Sleep,
	}
)

func GetScriptRuntimeMessageKey(nodeId, taskId string) string {
	return fmt.Sprintf("palm.nodebase.script.%v.%v", nodeId, taskId)
}

func GetNodeBaseNotificationQueueByNodeId(id string) string {
	return fmt.Sprintf("queue.notify-from-server.%v", id)
}

// Used to receive server-side notifications for nodes
func GetNodeBaseNotificationRoutingKeyByNodeId(id string) string {
	return fmt.Sprintf("palm.nodebase.notification.%v.#", id)
}

// Used to send server-side notifications for nodes
func GetServerPushKey(nodeId string, key ServerPushKey) string {
	return GetNodeBaseNotificationPushRoutingKeyByNodeId(nodeId, string(key))
}

func ParseServerPushKey(r string) string {
	rets := strings.Split(r, ".")
	if len(rets) >= 5 {
		return rets[4]
	}
	return ""
}

func GetNodeBaseNotificationPushRoutingKeyByNodeId(nodeId string, key string) string {
	return fmt.Sprintf("palm.nodebase.notification.%v.%v", nodeId, key)
}

func GetScanPortQueueNameByNodeId(nodeId string) string {
	return fmt.Sprintf("queue.scan-port-task.%v", nodeId)
}

func GetScanPortRoutingKeyByNodeId(nodeId string) string {
	return fmt.Sprintf("palm.stream.task.scan-port.%v", nodeId)
}

func GetScanFingerprintQueueNameByNodeId(nodeId string) string {
	return fmt.Sprintf("queue.scan-fingerprint-task.%v", nodeId)
}

func GetScanFingerprintRoutingKeyByNodeId(nodeId string) string {
	return fmt.Sprintf("palm.stream.task.scan-fingerprint.%v", nodeId)
}
