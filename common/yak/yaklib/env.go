package yaklib

import "os"

// Get to get the environment variable value of the corresponding key name
// ! Deprecated, you can use `os.Getenv` instead of
// Example:
// ```
// env.Get("PATH")
// ```
func _getEnv(key string) string {
	return os.Getenv(key)
}

// Set to set the environment variable value of the corresponding key name
// ! Deprecated, you can use `os.Setenv` instead of
// Example:
// ```
// env.Set("YAK_PROXY", "http://127.0.0.1:10808")
// ```
func _setEnv(key string, value string) {
	os.Setenv(key, value)
}

var EnvExports = map[string]interface{}{
	"Get": _getEnv,
	"Set": _setEnv,
}
