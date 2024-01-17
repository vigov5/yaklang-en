package yaklib

import (
	"time"

	"github.com/yaklang/yaklang/common/utils"
)

// now is used to obtain the current time. Time structure
// Example:
// ```
// dur = time.ParseDuration("1m")~
// ctx, cancel = context.WithDeadline(context.New(), now().Add(dur))
//
// println(now().Format("2006-01-02 15:04:05"))
// ```
func _timeNow() time.Time {
	return time.Now()
}

// now is used to obtain the current time. Time structure
// It is actually an alias of time.Now
// Example:
// ```
// dur = time.ParseDuration("1m")~
// ctx, cancel = context.WithDeadline(context.New(), now().Add(dur))
//
// println(now().Format("2006-01-02 15:04:05"))
// ```
func _timenow() time.Time {
	return time.Now()
}

// GetCurrentMonday Returns the time structure and error
// Example:
// ```
// monday, err = time.GetCurrentMonday()
// ```
func _getCurrentMonday() (time.Time, error) {
	return utils.GetCurrentWeekMonday()
}

// GetCurrentDate returns a time structure accurate to the current date with error
// Example:
// ```
// date, err = time.GetCurrentDate()
// ```
func _getCurrentDate() (time.Time, error) {
	return utils.GetCurrentDate()
}

// Parse and return the time structure and error
// is: 2006-01-02 15:04:05
// Example:
// ```
// t, err = time.Parse("2006-01-02 15:04:05", "2020-01-01 00:00:00")
// ```
func _timeparse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ParseDuration parses the time interval string according to the given format and returns the time interval structure and error
// time interval string is a possibly signed sequence of decimal numbers, each number can have optional decimal and unit suffixes, such as "300ms"，"-1.5h" or "2h45m"
// The valid time units are "ns"(nanoseconds), "us"(or "µs" (microseconds), "ms"(milliseconds), "s"（秒）, "m"（分）, "h"(hour)
// Example:
// ```
// d, err = time.ParseDuration("1h30m")
// ```
func _timeParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// The Unix function returns the corresponding local time structure based on the given Unix timestamp (sec seconds and nsec nanoseconds starting from UTC on January 1, 1970)
// Example:
// ```
// time.Unix(1577808000, 0) // 2020-01-01 00:00:00 +0800 CST
// ```
func _timeUnix(sec int64, nsec int64) time.Time {
	return time.Unix(sec, nsec)
}

// After is used to create a A timer that will send the current time to the returned channel after a specified time.
// Example:
// ```
// d, err = time.ParseDuration("5s")
// <-time.After(d) // to wait 5 seconds before executing subsequent statements.
// tln("after 5s")
// ```
func _timeAfter(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// AfterFunc is used to create a timer, which will execute the specified function after the specified time. The function will be executed in another coroutine.
// . The function itself will immediately return a timer structure reference. You can cancel the timer
// Example:
// ```
// d, err = time.ParseDuration("5s")
// timer = time.AfterFunc(d, () => println("after 5s")) // You can Cancel the timer by calling timer.Stop()
// time.sleep(10)
// ```
func _timeAfterFunc(d time.Duration, f func()) *time.Timer {
	return time.AfterFunc(d, f)
}

// NewTimer returns a timer structure reference based on the given time interval (unit: seconds)
// You can use <- timer.C to wait for the timer to expire
// . You can also call timer. Stop to cancel the timer
// Example:
// ```
// timer = time.NewTimer(5) // You can Cancel the timer by calling timer.Stop()
// <-timer.C // Wait for the timer to expire
// ```
func _timeNewTimer(d float64) *time.Timer {
	return time.NewTimer(utils.FloatSecondDuration(d))
}

// NewTicker returns a circular timer structure reference according to the given time interval (unit: seconds), which will periodically send the current time
// You can use <- timer.C to wait for the loop timer to expire
// You can also Cancel the loop timer by calling timer.Stop.
// Example:
// ```
// timer = time.NewTicker(5) // You can Cancel the timer by calling timer.Stop()
// ticker = time.NewTicker(1)
// for t in ticker.C {
// println("tick") // Prints a tick every 1 second
// }
// ```
func _timeNewTicker(d float64) *time.Ticker {
	return time.NewTicker(utils.FloatSecondDuration(d))
}

// The Until function returns the current time up to t (future time) time interval
// Example:
// ```
// t = time.Unix(1704038400, 0) // 2024-1-1 00:00:00 +0800 CST
// time.Until(t) // Returns the time interval from the current time to t
// ```
func _timeUntil(t time.Time) time.Duration {
	return time.Until(t)
}

// Since function returns the time interval from t (past time) to the current time
// Example:
// ```
// t = time.Unix(1577808000, 0) // 2020-01-01 00:00:00 +0800 CST
// time.Since(t) // Returns the time interval from t to the current time
// ```
func _timeSince(t time.Time) time.Duration {
	return time.Since(t)
}

var TimeExports = map[string]interface{}{
	"Now":              _timeNow,
	"now":              _timenow,
	"GetCurrentMonday": _getCurrentMonday,
	"GetCurrentDate":   _getCurrentDate,
	"sleep":            sleep,
	"Sleep":            sleep,
	"Parse":            _timeparse,
	"ParseDuration":    _timeParseDuration,
	"Unix":             _timeUnix,
	"After":            _timeAfter,
	"AfterFunc":        _timeAfterFunc,
	"NewTimer":         _timeNewTimer,
	"NewTicker":        _timeNewTicker,
	"Until":            _timeUntil,
	"Since":            _timeSince,
}
