package yaklib

import (
	"time"
	_ "time/tzdata"
)

// Get Returns the time zone with the given name
// If the name is an empty string "" or "UTC", LoadLocation returns the UTC time zone
// If the name is "Local"according to the time zone of the given name LoadLocation Returns the local time zone
// Otherwise, the name is treated as an IANA time zone A location name in the database, such as "America/New_York"
// Example:
// ```
// loc, err = timezone.Get("Asia/Shanghai")
// ```
func _timezoneLoadLocation(name string) (*time.Location, error) {
	return time.LoadLocation(name)
}

// Now Returns the current time structure
// Example:
// ```
// now = timezone.Now("Asia/Shanghai")
// ```
func _timezoneNow(name string) time.Time {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

var TimeZoneExports = map[string]interface{}{
	"Get": _timezoneLoadLocation,
	"Now": _timezoneNow,
}
