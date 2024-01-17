package yso

import (
	"github.com/yaklang/yaklang/common/facades/ldap/ldapserver"
)

var Exports = map[string]interface{}{
	// Generate chain
	"ToBytes": ToBytes,
	"ToBcel":  ToBcel,
	"ToJson":  ToJson,
	"dump":    Dump,
	//JavaObject
	"GetJavaObjectFromBytes":  GetJavaObjectFromBytes,
	"GetBeanShell1JavaObject": GetBeanShell1JavaObject,
	"GetClick1JavaObject":     GetClick1JavaObject,
	//"GetClojureJavaObject":                 GetClojureJavaObject,
	"GetCommonsBeanutils1JavaObject":       GetCommonsBeanutils1JavaObject,
	"GetCommonsBeanutils183NOCCJavaObject": GetCommonsBeanutils183NOCCJavaObject,
	"GetCommonsBeanutils192NOCCJavaObject": GetCommonsBeanutils192NOCCJavaObject,
	"GetCommonsCollections1JavaObject":     GetCommonsCollections1JavaObject,
	"GetCommonsCollections2JavaObject":     GetCommonsCollections2JavaObject,
	"GetCommonsCollections3JavaObject":     GetCommonsCollections3JavaObject,
	"GetCommonsCollections4JavaObject":     GetCommonsCollections4JavaObject,
	"GetCommonsCollections5JavaObject":     GetCommonsCollections5JavaObject,
	"GetCommonsCollections6JavaObject":     GetCommonsCollections6JavaObject,
	"GetCommonsCollections7JavaObject":     GetCommonsCollections7JavaObject,
	"GetCommonsCollections8JavaObject":     GetCommonsCollections8JavaObject,
	"GetCommonsCollectionsK1JavaObject":    GetCommonsCollectionsK1JavaObject,
	"GetCommonsCollectionsK2JavaObject":    GetCommonsCollectionsK2JavaObject,
	"GetCommonsCollectionsK3JavaObject":    GetCommonsCollectionsK3JavaObject,
	"GetCommonsCollectionsK4JavaObject":    GetCommonsCollectionsK4JavaObject,
	"GetGroovy1JavaObject":                 GetGroovy1JavaObject,
	"GetJBossInterceptors1JavaObject":      GetJBossInterceptors1JavaObject,
	"GetURLDNSJavaObject":                  GetURLDNSJavaObject,
	"GetFindGadgetByDNSJavaObject":         GetFindGadgetByDNSJavaObject,

	//"GetJRMPClientJavaObject":              GetJRMPClientJavaObject,
	"GetJSON1JavaObject":          GetJSON1JavaObject,
	"GetJavassistWeld1JavaObject": GetJavassistWeld1JavaObject,
	"GetJdk7u21JavaObject":        GetJdk7u21JavaObject,
	"GetJdk8u20JavaObject":        GetJdk8u20JavaObject,
	//Get gadgets in batches
	"GetAllGadget":            GetAllGadget,
	"GetAllTemplatesGadget":   GetAllTemplatesGadget,
	"GetAllRuntimeExecGadget": GetAllRuntimeExecGadget,
	//Get gadget name
	"GetGadgetNameByFun": GetGadgetNameByFun,
	//Used for Shiro check
	"GetSimplePrincipalCollectionJavaObject": GetSimplePrincipalCollectionJavaObject,
	// Load java class
	"LoadClassFromBytes":  LoadClassFromBytes,
	"LoadClassFromBase64": LoadClassFromBase64,
	"LoadClassFromBCEL":   LoadClassFromBCEL,

	// Only generate objects of malicious classes
	"GenerateClassObjectFromBytes":                     GenerateClassObjectFromBytes,
	"GenerateRuntimeExecEvilClassObject":               GenerateRuntimeExecEvilClassObject,
	"GenerateProcessBuilderExecEvilClassObject":        GenerateProcessBuilderExecEvilClassObject,
	"GenerateProcessImplExecEvilClassObject":           GenerateProcessImplExecEvilClassObject,
	"GenerateDNSlogEvilClassObject":                    GenDnslogClassObject,
	"GenerateSpringEchoEvilClassObject":                GenerateSpringEchoEvilClassObject,
	"GenerateModifyTomcatMaxHeaderSizeEvilClassObject": GenerateModifyTomcatMaxHeaderSizeEvilClassObject,
	"GenerateTcpReverseEvilClassObject":                GenTcpReverseClassObject,
	"GenerateTcpReverseShellEvilClassObject":           GenTcpReverseShellClassObject,
	"GenerateTomcatEchoClassObject":                    GenTomcatEchoClassObject,
	"GenerateMultiEchoClassObject":                     GenMultiEchoClassObject,
	"GenerateHeaderEchoClassObject":                    GenHeaderEchoClassObject,
	"GenerateSleepClassObject":                         GenSleepClassObject,
	// bytes class
	"useBytesEvilClass":         SetBytesEvilClass,
	"useBytesClass":             SetClassBytes,
	"useBase64BytesClass":       SetClassBase64Bytes,
	"useTomcatEchoEvilClass":    SetTomcatEchoEvilClass,
	"useTomcatEchoTemplate":     SetClassTomcatEchoTemplate,
	"useMultiEchoEvilClass":     SetMultiEchoEvilClass,
	"useClassMultiEchoTemplate": SetClassMultiEchoTemplate,
	//ModifyTomcatMaxHeaderSize
	"useModifyTomcatMaxHeaderSizeTemplate": SetClassModifyTomcatMaxHeaderSizeTemplate,
	//springecho template
	"useSpringEchoTemplate":   SetClassSpringEchoTemplate,
	"springHeader":            SetHeader,
	"springParam":             SetParam,
	"springRuntimeExecAction": SetExecAction,
	"springEchoBody":          SetEchoBody,
	// Dnslog template
	"useDNSlogTemplate":  SetClassDnslogTemplate,
	"dnslogDomain":       SetDnslog,
	"useDNSLogEvilClass": SetDnslogEvilClass,
	// runtime exec template
	"useRuntimeExecTemplate":  SetClassRuntimeExecTemplate,
	"command":                 SetExecCommand,
	"majorVersion":            SetMajorVersion,
	"useRuntimeExecEvilClass": SetRuntimeExecEvilClass,
	// runtime exec template
	"useProcessBuilderExecTemplate":  SetClassProcessBuilderExecTemplate,
	"useProcessBuilderExecEvilClass": SetProcessBuilderExecEvilClass,
	// runtime exec template
	"useProcessImplExecTemplate":  SetClassProcessImplExecTemplate,
	"useProcessImplExecEvilClass": SetProcessImplExecEvilClass,
	// tcp reverse template
	"useTcpReverseTemplate":  SetClassTcpReverseTemplate,
	"tcpReverseHost":         SetTcpReverseHost,
	"tcpReversePort":         SetTcpReversePort,
	"tcpReverseToken":        SetTcpReverseToken,
	"useTcpReverseEvilClass": SetTcpReverseEvilClass,
	// tcp reverse shell template
	"useTcpReverseShellTemplate":  SetClassTcpReverseShellTemplate,
	"useTcpReverseShellEvilClass": SetTcpReverseShellEvilClass,
	// header echo template
	"useHeaderEchoTemplate":  SetClassHeaderEchoTemplate,
	"useHeaderEchoEvilClass": SetHeaderEchoEvilClass,
	"useEchoBody":            SetEchoBody,
	"useParam":               SetParam,
	"useHeaderParam":         SetHeader,
	// sleep template
	"useSleepTemplate":  SetClassSleepTemplate,
	"useSleepEvilClass": SetSleepEvilClass,
	"useSleepTime":      SetSleepTime,
	// Other settings
	"useConstructorExecutor":       SetConstruct, // Use constructor to execute
	"evilClassName":                SetClassName, // className
	"obfuscationClassConstantPool": SetObfuscation,
}

var LDAPExports = map[string]interface{}{
	"NewLdapServer":         ldapserver.NewLdapServer,
	"NewLdapServerWithPort": ldapserver.NewLdapServerWithPort,
}
