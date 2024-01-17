package yaklib

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklang/spec"
	"github.com/yaklang/yaklang/common/yakdocument"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
)

// f is used to perform string manipulation. Formatted
// Example:
// ```
//
// str.f("hello %s", "yak") // hello yak
// ```
func _sfmt(f string, items ...interface{}) string {
	return fmt.Sprintf(f, items...)
}

// assert is used to determine whether the incoming Boolean value is true. If it is false, it will crash and print an error message.
// ! Deprecated, you can use the assert statement instead of
// Example:
// ```
// assert(code == 200, "code != 200") // . If the code is not equal to 200, it will crash and print an error message.
// // which is equivalent to assert code == 200, "code != 200"
// ```
func _assert(b bool, reason ...interface{}) {
	if !b {
		panic(spew.Sdump(reason))
	}
}

// assertf is used to determine whether the incoming Boolean value is true, and crashes if it is false And print the error message
// ! Deprecated, you can use the assert statement instead of
// Example:
// ```
// assertf(code == 200, "code != %d", 200) // . If the code is not equal to 200, it will crash and print an error message.
// // composed of n characters randomly selected from the uppercase and lowercase alphabet. It is equivalent to assert code == 200, sprintf("code != %d", 200)
// ```
func _assertf(b bool, f string, items ...any) {
	if !b {
		panic(_sfmt(f, items...))
	}
}

// assertEmpty is used to determine whether the incoming value is empty. If it is empty, it crashes and prints an error message.
// ! Deprecated, you can use the assert statement instead of
// Example:
// ```
// assertEmpty(nil, "nil is not empty") // . If nil is not empty, it will crash and print an error message. It will not crash here.
// ```
func _assertEmpty(i interface{}) {
	if i == nil || i == spec.Undefined {
		return
	}
	panic(_sfmt("expect nil but got %v", spew.Sdump(i)))
}

// fail crashes and prints an error message, which is actually almost equivalent to panic
// Example:
// ```
// try{
// 1/0
// } catch err {
// fail("exit code:", 1, "because:", err)
// }
// ```
func _failed(msg ...interface{}) {
	if msg == nil {
		panic("exit")
	}

	var msgs []string
	for _, i := range msg {
		if err, ok := i.(error); ok {
			msgs = append(msgs, err.Error())
		} else if s, ok := i.(string); ok {
			msgs = append(msgs, s)
		} else {
			msgs = append(msgs, spew.Sdump(i))
		}
	}
	panic(strings.Join(msgs, "\n"))
}

func yakitOutputHelper(i interface{}) {
	if yakitClientInstance != nil {
		yakitClientInstance.Output(i)
	}
}

// die determines whether the incoming error is empty. If not, If empty, it will crash and print an error message, which is actually equivalent to if err != nil { panic(err) }
// Example:
// ```
// die(err)
// ```
func _diewith(err interface{}) {
	if err == nil {
		return
	}
	yakitOutputHelper(fmt.Sprintf("YakVM Code DIE With Err: %v", spew.Sdump(err)))
	_failed(err)
}

// logdiscard. Use Used to discard all logs, that is, no longer display any logs.
// Example:
// ```
// logdiscard()
// ```
func _logDiscard() {
	log.SetOutput(io.Discard)
}

// logquiet is used to discard all logs, that is, no more logs are displayed. It is an alias of logdiscard.
// Example:
// ```
// logquiet()
// ```
func _logQuiet() {
	log.SetOutput(io.Discard)
}

// . logrecover is used to restore the display of logs. It is used to restore the effect caused by logdiscard.
// Example:
// ```
// logdiscard()
// logrecover()
// ```
func _logRecover() {
	log.SetOutput(os.Stdout)
}

func dummyN(items ...any) {
	if len(items) > 0 {
		fmt.Println(fmt.Sprintf(utils.InterfaceToString(items[0]), items[1:]...))
	}
}

// yakit_output is used to output logs in yakit. In the case of non-yakit, it will output the log in the console. When called in the mitm plug-in, the log will be output in the passive log.
// Example:
// ```
// yakit_output("hello %s", "yak")
// ```
func _yakit_output(items ...any) {
	if len(items) > 0 {
		fmt.Println(fmt.Sprintf(utils.InterfaceToString(items[0]), items[1:]...))
	}
}

// yakit_save
// ! Deprecated
func _yakit_save(items ...any) {
}

// yakit_status
// ! Deprecated
func _yakit_status(items ...any) {
}

// uuid is used to generate A uuid string
// Example:
// ```
// println(uuid())
// ```
func _uuid() string {
	return uuid.New().String()
}

// timestamp is used to obtain the current timestamp, and its return value is of type int64.
// Example:
// ```
// println(timestamp())
// ```
func _timestamp() int64 {
	return time.Now().Unix()
}

// nanotimestamp is used to obtain the current timestamp, accurate to nanoseconds, and its return value is int64 type
// Example:
// ```
// println(nanotimestamp())
// ```
func _nanotimestamp() int64 {
	return time.Now().UnixNano()
}

// date is used to obtain the current date in the format"2006-01-02“
// Example:
// ```
// println(date())
// ```
func _date() string {
	return time.Now().Format("2006-01-02")
}

// datetime is used to get the current date and time. Its format is"2006-01-02 15:04:05"
// Example:
// ```
// println(datetime())
// ```
func _datetime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// now is used to obtain the current time. Time structure
// is actually an alias of time.Now.
// Example:
// ```
// dur = time.ParseDuration("1m")~
// ctx, cancel = context.WithDeadline(context.New(), now().Add(dur))
//
// println(now().Format("2006-01-02 15:04:05"))
// ```
func _now() time.Time {
	return time.Now()
}

// timestampToDatetime is used to convert timestamps into dates and times in the format"2006-01-02 15:04:05"
// Example:
// ```
// println(timestampToDatetime(timestamp()))
// ```
func _timestampToDatetime(tValue int64) string {
	tm := time.Unix(tValue, 0)
	return tm.Format("2006-01-02 15:04:05")
}

// timestampToTime is used to convert the timestamp into a time structure
// Example:
// ```
// println(timestampToDatetime(timestamp()))
// ```
func _timestampToTime(tValue int64) time.Time {
	return time.Unix(tValue, 0)
}

// datetimeToTimestamp is used to convert date and time strings into timestamps, its format is"2006-01-02 15:04:05"
// Example:
// ```
// println(datetimeToTimestamp("2023-11-11 11:11:11")~)
// ```
func _datetimeToTimestamp(str string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// parseTime parses a formatted time string in a layout and returns the time structure it represents.
// Example:
// ```
// t, err = parseTime("2006-01-02 15:04:05", "2023-11-11 11:11:11")
// ```
func _parseTime(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// dump to format and print any type of data in a user-friendly way
// Example:
// ```
// dump("hello", 1, ["1", 2, "3"])
// ```
func _dump(i ...any) {
	spew.Dump(i...)
}

// sdump formats any type of data in a user-friendly way and returns the formatted string
// Example:
// ```
// println(sdump("hello", 1, ["1", 2, "3"]))
// ```
func _sdump(i ...any) string {
	return spew.Sdump(i...)
}

// randn is used to generate a Random number, its range is [min, max)
// If min is greater than max, an exception will be thrown
// Example:
// ```
// println(randn(1, 100))
// ```
func _randn(min, max int) int {
	if min > max {
		panic(_sfmt("randn failed; min: %v max: %v", min, max))
	}
	return min + rand.Intn(max-min)
}

// randstr returns a string
// Example:
// ```
// println(randstr(10))
// ```
func _randstr(length int) string {
	return utils.RandStringBytes(length)
}

// wait is used to wait for a context to complete, or to let the current coroutine sleep for a period of time. The unit is seconds
// Example:
// ```
// ctx, cancel = context.WithTimeout(context.New(), time.ParseDuration("5s")~) // The context is completed after calling the cancel function or 5 seconds
// wait(ctx) // waits for the context to complete.
// wait(1.5) // Sleep for 1.5 seconds
// ```
func _wait(i interface{}) {
	switch ret := i.(type) {
	case context.Context:
		select {
		case <-ret.Done():
		}
	case string:
		sleep(parseFloat(ret))
	case float64:
		sleep(ret)
	case float32:
		sleep(float64(ret))
	case int:
		sleep(float64(ret))
	default:
		panic(fmt.Sprintf("cannot wait %v", spew.Sdump(ret)))
	}
}

// isEmpty is used to determine the transfer Whether the input value is empty, if it is empty, return true, otherwise return false
// Example:
// ```
// isEmpty(nil) // true
// isEmpty(1) // false
// ```
func _isEmpty(i interface{}) bool {
	if i == nil || i == spec.Undefined {
		return true
	}
	return false
}

// chr. Converts the incoming value to the corresponding character according to the ascii code table.
// Example:
// ```
// chr(65) // A
// chr("65") // A
// ```
func chr(i interface{}) string {
	switch v := i.(type) {
	case int:
		return string(rune(v))
	case int8:
		return string(rune(v))
	case int16:
		return string(rune(v))
	case int32:
		return string(rune(v))
	case int64:
		return string(rune(v))
	case uint:
		return string(rune(v))
	case uint8:
		return string(rune(v))
	case uint16:
		return string(rune(v))
	case uint32:
		return string(rune(v))
	case uint64:
		return string(rune(v))
	default:
		return string(rune(parseInt(utils.InterfaceToString(i))))
	}
}

// ord converts the incoming value into the corresponding ascii code integer
// Example:
// ```
// ord("A") // 65
// ord('A') // 65
// ```
func ord(i interface{}) int {
	switch ret := i.(type) {
	case byte:
		return int(ret)
	default:
		strRaw := utils.InterfaceToString(i)
		if strRaw == "" {
			return -1
		}

		if r := []rune(strRaw); r != nil {
			return int(r[0])
		}

		return int(strRaw[0])
	}
}

// typeof is used to get the incoming The value type structure
// Example:
// ```
// typeof(1) == int // true
// typeof("hello") == string // true
// ```
func typeof(i interface{}) reflect.Type {
	return reflect.TypeOf(i)
}

// desc Print the details of the incoming complex value in a user-friendly way, which is often a structure or a structure reference. Detailed information includes available fields and available member methods.
// Example:
// ```
// ins = fuzz.HTTPRequest(poc.BasicRequest())~
// desc(ins)
// ```
func _desc(i interface{}) {
	info, err := yakdocument.Dir(i)
	if err != nil {
		log.Error(err)
		return
	}
	if info == nil {
		return
	}
	info.Show()
}

// descStr. Prints the detailed information of the incoming complex value in a user-friendly way, which is often a structure or a structure reference. The detailed information includes available fields, available Member method, returns a string of detailed information
// Example:
// ```
// ins = fuzz.HTTPRequest(poc.BasicRequest())~
// println(descStr(ins))
// ```
func _descToString(i interface{}) string {
	info, err := yakdocument.Dir(i)
	if err != nil {
		log.Error(err)
		return ""
	}
	if info == nil {
		return ""
	}
	return info.String()
}

// tick1s is used to execute the incoming function every 1 second until the function returns false.
// Example:
// ```
// count = 0
// tick1s(func() bool {
// println("hello")
// count++
// return count <= 5
// })
func tick1s(f func() bool) {
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-t.C:
			if !f() {
				return
			}
		}
	}
}

// sleep is used to make the current coroutine sleep for a period of time. The unit is seconds.
// Example:
// ```
// sleep(1.5) // Sleep for 1.5 seconds
// ```
func sleep(i float64) {
	time.Sleep(utils.FloatSecondDuration(i))
}

var GlobalExport = map[string]interface{}{
	"_createOnLogger":        createLogger,
	"_createOnLoggerConsole": createConsoleLogger,
	"_createOnFailed":        createFailed,
	"_createOnOutput":        createOnOutput,
	"_createOnFinished":      createOnFinished,
	"_createOnAlert":         createOnAlert,

	"loglevel":   setLogLevel,
	"logquiet":   _logDiscard,
	"logdiscard": _logDiscard,
	"logrecover": _logRecover,

	"yakit_output": _yakit_output,
	"yakit_save":   _yakit_save,
	"yakit_status": _yakit_status,

	"fail": _failed,
	"die":  _diewith,
	"uuid": _uuid,

	"timestamp":           _timestamp,
	"nanotimestamp":       _nanotimestamp,
	"datetime":            _datetime,
	"date":                _date,
	"now":                 _now,
	"parseTime":           _parseTime,
	"timestampToDatetime": _timestampToDatetime,
	"timestampToTime":     _timestampToTime,
	"datetimeToTimestamp": _datetimeToTimestamp,
	"tick1s":              tick1s,

	"input": _input,
	"dump":  _dump,
	"sdump": _sdump,

	"randn":   _randn,
	"randstr": _randstr,

	"assert":      _assert,
	"assertTrue":  _assert,
	"isEmpty":     _isEmpty,
	"assertEmpty": _assertEmpty,
	"assertf":     _assertf,

	"parseInt":     parseInt,
	"parseStr":     parseString,
	"parseString":  parseString,
	"parseBool":    parseBool,
	"parseBoolean": parseBool,
	"parseFloat":   parseFloat,
	"atoi":         atoi,

	"sleep": sleep,
	"wait":  _wait,

	"desc":    _desc,
	"descStr": _descToString,
	"chr":     chr,
	"ord":     ord,
	"type":    typeof,
	"typeof":  typeof,
}
