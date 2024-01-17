package yaklib

import (
	"context"
	"time"

	"github.com/yaklang/yaklang/common/utils"
)

// Seconds Returns a Context interface with a timeout of d seconds (i.e. context interface)
// It is actually an alias of context.WithTimeoutSeconds
// Example:
// ```
// ctx = context.Seconds(10)
// ```
func _seconds(d float64) context.Context {
	return utils.TimeoutContextSeconds(d)
}

// WithTimeoutSeconds returns a Context interface with a timeout of d seconds (i.e. context interface)
// Example:
// ```
// ctx = context.WithTimeoutSeconds(10)
// ```
func _withTimeoutSeconds(d float64) context.Context {
	return utils.TimeoutContextSeconds(d)
}

// New returns an empty Context interface (i.e. context interface)
// It is actually an alias of context.Background
// Example:
// ```
// ctx = context.New()
// ```
func _newContext() context.Context {
	return context.Background()
}

// Background returns an empty Context interface (i.e. context interface)
// Example:
// ```
// ctx = context.Background()
// ```
func _background() context.Context {
	return context.Background()
}

// WithCancel returns the Context interface (i.e. context interface) inherited from parent and the cancellation function
// . When the returned cancellation function or the parents cancellation function is called, the entire context will be canceled.
// Example:
// ```
// ctx, cancel = context.WithCancel(context.Background())
// defer cancel()
// ```
func _withCancel(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(parent)
}

// WithTimeout returns the Context interface inherited from parent (i.e. context interface ) and the cancellation function
// when calling the returned cancellation function or timeout, the entire context will be canceled
// Example:
// ```
// dur, err = time.ParseDuration("10s")
// ctx, cancel := context.WithTimeout(context.Background(), dur)
// defer cancel()
// ```
func _withTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// WithDeadline returns the Context interface inherited from parent (i.e. context interface) and the cancellation function
// When the call returns The cancellation function or the specified time is exceeded, the entire context will be canceled
// Example:
// ```
// dur, err = time.ParseDuration("10s")
// after = time.Now().Add(dur)
// ctx, cancel := context.WithDeadline(context.Background(), after)
// defer cancel()
// ```
func _withDeadline(parent context.Context, t time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, t)
}

// WithValue returns the Context interface (that is, the context interface) that inherits the key value and carries the additional key value and the cancellation function
// When the returned cancellation function is called, the entire context will be canceled
// Example:
// ```
// ctx = context.WithValue(context.Background(), "key", "value")
// ctx.Value("key") // "value"
// ```
func _withValue(parent context.Context, key, val any) context.Context {
	return context.WithValue(parent, key, val)
}

var ContextExports = map[string]interface{}{
	"Seconds":            _seconds,
	"New":                _newContext,
	"Background":         _background,
	"WithCancel":         _withCancel,
	"WithTimeout":        _withTimeout,
	"WithTimeoutSeconds": _withTimeoutSeconds,
	"WithDeadline":       _withDeadline,
	"WithValue":          _withValue,
}
