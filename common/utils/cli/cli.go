package cli

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
)

type cliExtraParams struct {
	optName      string
	optShortName string
	params       []string
	defaultValue interface{}
	helpInfo     string
	required     bool
	tempArgs     []string
	_type        string
}

var (
	Args []string

	cliParamInvalid = utils.NewBool(false)
	cliName         = "cmd"
	cliDocument     = ""
	errorMsg        = ""

	helpParam = &cliExtraParams{
		optShortName: "h",
		optName:      "help",
		params:       []string{"-h", "--help"},
		defaultValue: false,
		helpInfo:     "Show help information",
		required:     false,
		_type:        "bool",
	}

	currentExtraParams []*cliExtraParams = []*cliExtraParams{
		helpParam,
	}
)

func init() {
	Args = os.Args[:]
	if len(Args) > 1 {
		filename := filepath.Base(os.Args[1])
		fileSuffix := path.Ext(filename)
		cliName = strings.TrimSuffix(filename, fileSuffix)
	}
}

func InjectCliArgs(args []string) {
	Args = args
}

func (param *cliExtraParams) foundArgsIndex() int {
	args := _getArgs()
	if param.tempArgs != nil {
		args = param.tempArgs
	}
	for _, opt := range param.params {
		if ret := utils.StringArrayIndex(args, opt); ret < 0 {
			continue
		} else {
			return ret
		}
	}
	return -1
}

func (param *cliExtraParams) GetDefaultValue(i interface{}) interface{} {
	if param.defaultValue != nil {
		return param.defaultValue
	}

	if !param.required {
		return i
	}

	cliParamInvalid.Set()
	errorMsg += fmt.Sprintf("\n  Parameter [%s] error: miss parameter", param.optName)
	return i
}

type SetCliExtraParam func(c *cliExtraParams)

// help Used to output help information for the command line program
// Example:
// ```
// cli.help()
// ```
func _help(w ...io.Writer) {
	var writer io.Writer = os.Stdout

	if len(w) > 0 {
		writer = w[0]
	}

	fmt.Fprintln(writer, "Usage: ")
	fmt.Fprintf(writer, "  %s [OPTIONS]\n", cliName)
	fmt.Fprintln(writer)
	if len(cliDocument) > 0 {
		fmt.Fprintln(writer, cliDocument)
		fmt.Fprintln(writer)
	}
	fmt.Fprintln(writer, "Flags:")
	for _, param := range currentExtraParams {
		paramType := param._type
		// bool type does not display paramType
		if paramType == "bool" {
			paramType = ""
		}
		helpInfo := param.helpInfo
		if param.defaultValue != nil {
			helpInfo += fmt.Sprintf(" (default %v)", param.defaultValue)
		}
		flag := fmt.Sprintf("  %s %s", strings.Join(param.params, ", "), paramType)
		padding := ""
		if len(flag) < 30 {
			padding = strings.Repeat(" ", 30-len(flag))
		}

		fmt.Fprintf(writer, "%v%v%v\n", flag, padding, param.helpInfo)
	}
}

// check is used to check whether the command line parameters are legal, which mainly checks the necessary parameters Whether the input and the input value are legal
// Example:
// ```
// target = cli.String("target", cli.SetRequired(true))
// cli.check()
// ```
func _cliCheck() {
	if helpParam.foundArgsIndex() != -1 {
		_help()
		os.Exit(1)
	} else if cliParamInvalid.IsSet() {
		errorMsg = strings.TrimSpace(errorMsg)
		if len(errorMsg) > 0 {
			fmt.Printf("Error:\n  %s\n\n", errorMsg)
		}
		_help()

	}
}

func CliCheckWithContext(cancel context.CancelFunc) func() {
	return func() {
		if helpParam.foundArgsIndex() != -1 {
			_help()
			cancel()
		} else if cliParamInvalid.IsSet() {
			errorMsg = strings.TrimSpace(errorMsg)
			if len(errorMsg) > 0 {
				fmt.Printf("Error:\n  %s\n\n", errorMsg)
			}
			_help()

		}
	}
}

// SetCliName Set the name of this command line program
// This will display
// Example:
// ```
// cli.SetCliName("example-tools")
// ```
func _cliSetName(name string) {
	cliName = name
}

// SetDoc Set the document of this command line program
// This will be entered on the command line --help or execute `cli.check()` after the parameter When it is illegal,
// Example:
// ```
// cli.SetDoc("example-tools is a tool for example")
// ```
func _cliSetDocument(document string) {
	cliDocument = document
}

// setDefaultValue is an option function, sets the default value of the parameter
// Example:
// ```
// cli.String("target", cli.SetDefaultValue("yaklang.com"))
// ```
func _cliSetDefaultValue(i interface{}) SetCliExtraParam {
	return func(c *cliExtraParams) {
		c.defaultValue = i
	}
}

// setHelp is an option function, set the help information of the parameter
// This will be entered on the command line --help or execute `cli.check()` after the parameter When it is illegal,
// Example:
// ```
// cli.String("target", cli.SetHelp("target host or ip"))
// ```
func _cliSetHelpInfo(i string) SetCliExtraParam {
	return func(c *cliExtraParams) {
		c.helpInfo = i
	}
}

func SetTempArgs(args []string) SetCliExtraParam {
	return func(c *cliExtraParams) {
		c.tempArgs = args
	}
}

// setRequired is an option. Function, set whether the parameters must be
// Example:
// ```
// cli.String("target", cli.SetRequired(true))
// ```
func _cliSetRequired(t bool) SetCliExtraParam {
	return func(c *cliExtraParams) {
		c.required = t
	}
}

func _getExtraParams(name string, opts ...SetCliExtraParam) *cliExtraParams {
	optName := name
	optShortName := ""
	if strings.Contains(name, " ") {
		nameSlice := strings.SplitN(name, " ", 2)
		optShortName = nameSlice[0]
		optName = nameSlice[1]
	} else if strings.Contains(name, ",") {
		nameSlice := strings.SplitN(name, ",", 2)
		optShortName = nameSlice[0]
		optName = nameSlice[1]
	}

	if len(name) == 1 && optShortName == "" {
		optShortName = name
		optName = name
	}

	param := &cliExtraParams{
		optName:      optName,
		optShortName: optShortName,
		params:       _getAvailableParams(optName, optShortName),
		required:     false,
		defaultValue: nil,
		helpInfo:     "",
	}
	for _, opt := range opts {
		opt(param)
	}
	currentExtraParams = append(currentExtraParams, param)
	return param
}

// ------------------------------------------------

func _getAvailableParams(optName, optShortName string) []string {
	optName, optShortName = strings.TrimLeft(optName, "-"), strings.TrimLeft(optShortName, "-")

	if optShortName == "" {
		return []string{fmt.Sprintf("--%v", optName)}
	}
	return []string{fmt.Sprintf("-%v", optShortName), fmt.Sprintf("--%v", optName)}
}

// Args.
// Example:
// ```
// Args = cli.Args()
// ```
func _getArgs() []string {
	return Args
}

func _cliFromString(name string, opts ...SetCliExtraParam) (string, *cliExtraParams) {
	param := _getExtraParams(name, opts...)
	index := param.foundArgsIndex()
	if index < 0 {
		return utils.InterfaceToString(param.GetDefaultValue("")), param
	}
	args := _getArgs()
	if param.tempArgs != nil {
		args = param.tempArgs
	}
	if index+1 >= len(args) {
		// prevents the array from going out of bounds
		return utils.InterfaceToString(param.GetDefaultValue("")), param
	}
	return args[index+1], param
}

// Bool Get the command line parameter of the corresponding name and convert it to bool type return
// Example:
// ```
// verbose = cli.Bool("verbose") // --verbose is true
// ```
func _cliBool(name string, opts ...SetCliExtraParam) bool {
	c := _getExtraParams(name, opts...)
	c._type = "bool"
	c.required = false

	index := c.foundArgsIndex()
	if index < 0 {
		return false // c.GetDefaultValue(false).(bool)
	}
	return true
}

// Int Gets the command line parameter of the corresponding name and converts it to int type Returns
// Example:
// ```
// target = cli.String("target") // --port 80 then the port is 80
// ```
func CliString(name string, opts ...SetCliExtraParam) string {
	s, c := _cliFromString(name, opts...)
	c._type = "string"
	return s
}

func parseInt(s string) int {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Errorf("parse int[%s] failed: %s", s, err)
		return 0
	}
	return int(i)
}

// Int gets the command line parameter of the corresponding name and converts it to int type returns
// Example:
// ```
// port = cli.Int("port") // --port 80 then the port is 80
// ```
func _cliInt(name string, opts ...SetCliExtraParam) int {
	s, c := _cliFromString(name, opts...)
	c._type = "int"
	if s == "" {
		return 0
	}
	return parseInt(s)
}

func parseFloat(s string) float64 {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Errorf("parse float[%s] failed: %s", s, err)
		return 0
	}
	return float64(i)
}

// Float gets the command line parameter of the corresponding name and converts it to float type. Returns
// Example:
// ```
// percent = cli.Float("percent") // --percent 0.5 Then the percent is 0.5
func _cliFloat(name string, opts ...SetCliExtraParam) float64 {
	s, c := _cliFromString(name, opts...)
	c._type = "float"
	if s == "" {
		return 0.0
	}
	return parseFloat(s)
}

// Urls Get the command line parameters of the corresponding name, according to","cut and try to convert it into a URL format and return []string type
// Example:
// ```
// urls = cli.Urls("urls")
// // --targets yaklang.com,google.com Then targets are ["https://yaklang.com", "https://google.com"]
// ```
func _cliUrls(name string, opts ...SetCliExtraParam) []string {
	s, c := _cliFromString(name, opts...)
	c._type = "urls"
	ret := utils.ParseStringToUrlsWith3W(utils.ParseStringToHosts(s)...)
	if ret == nil {
		return []string{}
	}
	return ret
}

// Ports Get the command line parameters of the corresponding name, set the parameter group name according to","and"-"cut and try to parse the port and return [ ]int type
// Example:
// ```
// ports = cli.Ports("ports")
// // --ports 10086-10088,23333, then ports are [10086, 10087, 10088, 23333]
// ```
func _cliPort(name string, opts ...SetCliExtraParam) []int {
	s, c := _cliFromString(name, opts...)
	c._type = "port"
	ret := utils.ParseStringToPorts(s)
	if ret == nil {
		return []int{}
	}
	return ret
}

// Hosts get the command line parameters of the corresponding name. According to","and try to parse the CIDR network segment and return []string type
// Example:
// ```
// hosts = cli.Hosts("hosts")
// // --hosts 192.168.0.0/24,172.17.0.1 then hosts is 192.168.0.0/. All IPs corresponding to 24 and 172.17.0.1
func _cliHosts(name string, opts ...SetCliExtraParam) []string {
	s, c := _cliFromString(name, opts...)
	c._type = "hosts"
	ret := utils.ParseStringToHosts(s)
	if ret == nil {
		cliParamInvalid.Set()
		errorMsg += fmt.Sprintf("\n  Parameter [%s] error: Parse string to host error: %s", c.optName, s)
		return []string{}
	}
	return ret
}

// FileOrContent Gets the command line parameter of the corresponding name
// Example:
// ```
// file = cli.File("file")
// // --file /etc/passwd then file is /etc/passwd file
// ```
func _cliFile(name string, opts ...SetCliExtraParam) []byte {
	s, c := _cliFromString(name, opts...)
	c._type = "file"
	c.required = true

	if cliParamInvalid.IsSet() {
		return []byte{}
	}

	if utils.GetFirstExistedPath(s) == "" && !cliParamInvalid.IsSet() {
		cliParamInvalid.Set()
		errorMsg += fmt.Sprintf("\n  Parameter [%s] error: No such file: %s", c.optName, s)
		return []byte{}
	}
	raw, err := ioutil.ReadFile(s)
	if err != nil {
		cliParamInvalid.Set()
		errorMsg += fmt.Sprintf("\n  Parameter [%s] error: %s", c.optName, err.Error())
		return []byte{}
	}

	return raw
}

// FileNames obtains the command line parameters of the corresponding names, obtains all selected file paths, and returns the []string type.
// Example:
// ```
// file = cli.FileNames("file")
// // --file /etc/passwd,/etc/hosts, then file is ["/etc/passwd", "/etc/hosts"]
// ```
func _cliFileNames(name string, opts ...SetCliExtraParam) []string {
	rawStr, c := _cliFromString(name, opts...)
	c._type = "file-names"

	if rawStr == "" {
		return []string{}
	}

	return utils.PrettifyListFromStringSplited(rawStr, ",")
}

// FileOrContent Get the command line with the corresponding name Parameters
// attempts to read the corresponding file content based on the passed in value. If it cannot be read, it returns directly, and finally returns []byte. Type
// Example:
// ```
// foc = cli.FileOrContent("foc")
// // --foc /etc/passwd then foc is /etc/passwd file
// // --file "asd" then file is "asd"
// ```
func _cliFileOrContent(name string, opts ...SetCliExtraParam) []byte {
	s, c := _cliFromString(name, opts...)
	c._type = "file_or_content"
	ret := utils.StringAsFileParams(s)
	if ret == nil {
		cliParamInvalid.Set()
		errorMsg += fmt.Sprintf("\n  Parameter [%s] error: Empty file or content: %s", c.optName, s)
		return []byte{}
	}
	return ret
}

// LineDict gets the command line parameter of the corresponding name
// and tries to read its corresponding file based on the value passed in Content, if it cannot be read, it will be used as a string, and finally cut according to the newline character, return []string type
// Example:
// ```
// dict = cli.LineDict("dict")
// // --dict /etc/passwd Then the dict is /etc/passwd file
// // --dict "asd" then the dict is ["asd"]
// ```
func _cliLineDict(name string, opts ...SetCliExtraParam) []string {
	s, c := _cliFromString(name, opts...)
	c._type = "file-or-content"
	raw := utils.StringAsFileParams(s)
	if raw == nil {
		cliParamInvalid.Set()
		errorMsg += fmt.Sprintf("\n  Parameter [%s] error: Empty file or content: %s", c.optName, s)
		return []string{}
	}

	return utils.ParseStringToLines(string(raw))
}

// named yakit-plugin-file. This will display
// according to the value passed in. Read the corresponding file content and based on"|"cuts and returns the []string type. Indicates the name of each plug-in
// Example:
// ```
// plugins = cli.YakitPlugin()
// // --yakit-plugin-file plugins.txt then plugins is plugins Each plug-in name in the .txt file
// ```
func _cliYakitPluginFiles(options ...SetCliExtraParam) []string {
	paramName := "yakit-plugin-file"
	filename, c := _cliFromString(paramName, options...)
	c._type = "yakit-plugin"

	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		cliParamInvalid.Set()
		errorMsg += fmt.Sprintf("\n  Parameter [%s] error: %s", c.optName, err.Error())
		return []string{}
	}
	if raw == nil {
		cliParamInvalid.Set()
		errorMsg += fmt.Sprintf("\n  Parameter [%s] error: Can't read file: %s", c.optName, filename)
		return []string{}
	}
	return utils.PrettifyListFromStringSplited(string(raw), "|")
}

// StringSlice Get the command line parameters of the corresponding name, cut the string according to","cutting returns []string type
// Example:
// ```
// targets = cli.StringSlice("targets")
// // --targets yaklang.com,google.com then targets is ["yaklang.com", "google.com"]
// ```
func CliStringSlice(name string, options ...SetCliExtraParam) []string {
	rawStr, c := _cliFromString(name, options...)
	c._type = "string-slice"

	if rawStr == "" {
		return []string{}
	}

	return utils.PrettifyListFromStringSplited(rawStr, ",")
}

// setVerboseName is an option function, set the Chinese name of the parameter
// Example:
// ```
// cli.String("target", cli.setVerboseName("target"))
// ```
func _cliSetVerboseName(verboseName string) {
}

// setGroup Yes An option function, set the parameter group
// Example:
// ```
// cli.String("target", cli.setGroup("common"))
// cli.Int("port", cli.setGroup("common"))
// cli.Int("threads", cli.setGroup("request"))
// cli.Int("retryTimes", cli.setGroup("request"))
// ```
func _cliSetGroup(group string) {
}

// setMultiSelect is an option function, Set whether the parameter can be multi-selected
// This option is only effective in `cli.StringSlice`
// Example:
// ```
// cli.StringSlice("targets", cli.setMultiSelect(true))
// ```
func _cliSetMultiSelect(multiSelect bool) {
}

// setSelectOption is an option function, set the parameter drop-down box option
// This option is only effective in `cli.StringSlice`
// Example:
// ```
// cli.StringSlice("targets", cli.setSelectOption("Drop-down box option", "drop-down box value"))
// ```
func _cliSetSelectOption(name, value string) {
}

var CliExports = map[string]interface{}{
	"Args":        _getArgs,
	"Bool":        _cliBool,
	"Have":        _cliBool,
	"String":      CliString,
	"HTTPPacket":  CliString,
	"YakCode":     CliString,
	"Text":        CliString,
	"Int":         _cliInt,
	"Integer":     _cliInt,
	"Float":       _cliFloat,
	"Double":      _cliFloat,
	"YakitPlugin": _cliYakitPluginFiles,
	"StringSlice": CliStringSlice,

	// parses into URL
	"Urls": _cliUrls,
	"Url":  _cliUrls,

	// parsing port
	"Ports": _cliPort,
	"Port":  _cliPort,

	// parses the network target
	"Hosts":   _cliHosts,
	"Host":    _cliHosts,
	"Network": _cliHosts,
	"Net":     _cliHosts,

	// Parse files and other
	"File":          _cliFile,
	"FileNames":     _cliFileNames,
	"FileOrContent": _cliFileOrContent,
	"LineDict":      _cliLineDict,

	// Set the param attribute
	"setHelp":     _cliSetHelpInfo,
	"setDefault":  _cliSetDefaultValue,
	"setRequired": _cliSetRequired,
	// set the Chinese name
	"setVerboseName": _cliSetVerboseName,
	// 设置参数组名
	"setCliGroup": _cliSetGroup,
	// set whether to multi-select (only supports `cli.StringSlice`)
	"setMultipleSelect": _cliSetMultiSelect,
	// Set the drop-down box option (only supports `cli.StringSlice`)
	"setSelectOption": _cliSetSelectOption,

	// Set the cli attribute
	"SetCliName": _cliSetName,
	"SetDoc":     _cliSetDocument,

	// general function
	"help":  _help,
	"check": _cliCheck,
}
