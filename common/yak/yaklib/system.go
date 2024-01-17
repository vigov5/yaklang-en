package yaklib

import (
	"net"
	"os"
	"runtime"

	"github.com/yaklang/yaklang/common/netx"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/cli"
	"github.com/yaklang/yaklang/common/utils/privileged"
)

var SystemExports = map[string]interface{}{
	"IsTCPPortOpen":             IsTCPPortOpen,
	"IsUDPPortOpen":             IsUDPPortOpen,
	"LookupHost":                lookupHost,
	"LookupIP":                  lookupIP,
	"IsTCPPortAvailable":        IsTCPPortAvailable,
	"IsUDPPortAvailable":        IsUDPPortAvailable,
	"GetRandomAvailableTCPPort": GetRandomAvailableTCPPort,
	"GetRandomAvailableUDPPort": GetRandomAvailableUDPPort,
	"IsRemoteTCPPortOpen":       IsRemoteTCPPortOpen,

	"GetMachineID": GetMachineID,

	"Remove":               Remove,
	"RemoveAll":            RemoveAll,
	"Rename":               Rename,
	"TempDir":              TempDir,
	"Getwd":                Getwd,
	"Getpid":               Getpid,
	"Getppid":              Getppid,
	"Getuid":               Getuid,
	"Geteuid":              Geteuid,
	"Getgid":               Getgid,
	"Getegid":              Getegid,
	"Environ":              Environ,
	"Hostname":             Hostname,
	"Unsetenv":             Unsetenv,
	"LookupEnv":            LookupEnv,
	"Clearenv":             Clearenv,
	"Setenv":               Setenv,
	"Getenv":               Getenv,
	"Exit":                 Exit,
	"Args":                 cli.Args,
	"Stdout":               Stdout,
	"Stdin":                Stdin,
	"Stderr":               Stderr,
	"Executable":           Executable,
	"ExpandEnv":            ExpandEnv,
	"Pipe":                 Pipe,
	"Chdir":                Chdir,
	"Chmod":                Chmod,
	"Chown":                Chown,
	"OS":                   OS,
	"ARCH":                 ARCH,
	"IsPrivileged":         IsPrivileged,
	"GetDefaultDNSServers": GetDefaultDNSServers,
	"WaitConnect":          WaitConnect,
	"GetLocalAddress":      GetLocalAddress,
	"GetLocalIPv4Address":  GetLocalIPv4Address,
	"GetLocalIPv6Address":  GetLocalIPv6Address,
}

// LookupHost Find the IP based on the domain name through the DNS server
// Example:
// ```
// os.LookupHost("www.yaklang.com")
// ```
func lookupHost(i string) []string {
	return netx.LookupAll(i)
}

// LookupIP Find the IP based on the domain name through the DNS server
// Example:
// ```
// os.LookupIP("www.yaklang.com")
// ```
func lookupIP(i string) []string {
	return netx.LookupAll(i)
}

// IsTCPPortOpen Check whether the TCP port is open
// Example:
// ```
// os.IsTCPPortOpen(80)
// ```
func IsTCPPortOpen(p int) bool {
	return !utils.IsTCPPortAvailable(p)
}

// IsUDPPortOpen Check whether the UDP port is open
// Example:
// ```
// os.IsUDPPortOpen(80)
// ```
func IsUDPPortOpen(p int) bool {
	return !utils.IsUDPPortAvailable(p)
}

// IsTCPPortAvailable Check whether the TCP port is available
// Example:
// ```
// os.IsTCPPortAvailable(80)
// ```
func IsTCPPortAvailable(p int) bool {
	return utils.IsTCPPortAvailable(p)
}

// IsUDPPortAvailable Check whether the UDP port is available
// Example:
// ```
// os.IsUDPPortAvailable(80)
// ```
func IsUDPPortAvailable(p int) bool {
	return utils.IsUDPPortAvailable(p)
}

// GetRandomAvailableTCPPort Gets the randomly available TCP port
// Example:
// ```
// tcp.Serve("127.0.0.1", os.GetRandomAvailableTCPPort())
// ```
func GetRandomAvailableTCPPort() int {
	return utils.GetRandomAvailableTCPPort()
}

// GetRandomAvailableUDPPort Get a randomly available UDP port
// Example:
// ```
// udp.Serve("127.0.0.1", os.GetRandomAvailableTCPPort())
// ```
func GetRandomAvailableUDPPort() int {
	return utils.GetRandomAvailableUDPPort()
}

// IsRemoteTCPPortOpen Check whether the remote TCP port is open
// Example:
// ```
// os.IsRemoteTCPPortOpen("yaklang.com", 443) // true
// ```
func IsRemoteTCPPortOpen(host string, p int) bool {
	return utils.IsTCPPortOpen(host, p)
}

// GetMachineID Gets the unique identifier of each machine
// Example:
// ```
// os.GetMachineID()
// ```
func GetMachineID() string {
	return utils.GetMachineCode()
}

// Remove Delete the specified file or directory
// Example:
// ```
// os.Remove("/tmp/test.txt")
// ```
func Remove(name string) error {
	return os.Remove(name)
}

// RemoveAll Recursively delete the specified path and its subpaths
// Example:
// ```
// os.RemoveAll("/tmp")
// ```
func RemoveAll(name string) error {
	return os.RemoveAll(name)
}

// Rename Renames a file or directory, which can be used to move files or directories
// Example:
// ```
// os.Rename("/tmp/test.txt", "/tmp/test2.txt")
// os.Rename("/tmp/test", "/root/test")
// ```
func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// TempDir Get the default directory path used to store temporary files
// Example:
// ```
// os.TempDir()
// ```
func TempDir() string {
	return os.TempDir()
}

// Getwd Get the current working directory path
// Example:
// ```
// cwd, err = os.Getwd()
// ```
func Getwd() (string, error) {
	return os.Getwd()
}

// Getpid Gets the process ID of the current process
// Example:
// ```
// os.Getpid()
// ```
func Getpid() int {
	return os.Getpid()
}

// Getppid Gets the parent process ID of the current process
// Example:
// ```
// os.Getppid()
// ```
func Getppid() int {
	return os.Getppid()
}

// Getuid Gets the user ID of the current process
// Example:
// ```
// os.Getuid()
// ```
func Getuid() int {
	return os.Getuid()
}

// Geteuid Get the effective user ID of the current process
// Example:
// ```
// os.Geteuid()
// ```
func Geteuid() int {
	return os.Geteuid()
}

// Getgid Get the group ID of the current process
// Example:
// ```
// os.Getgid()
// ```
func Getgid() int {
	return os.Getgid()
}

// Getegid Gets the effective group ID of the current process
// Example:
// ```
// os.Getegid()
// ```
func Getegid() int {
	return os.Getegid()
}

// Environ Gets the string slice representing the environment variable in the format"key=value"
// Example:
// ```
// for env in os.Environ() {
// value = env.SplitN("=", 2)
// printf("key = %s, value = %v\n", value[0], value[1])
// }
// ```
func Environ() []string {
	return os.Environ()
}

// Hostname Get the host name
// Example:
// ```
// name, err = os.Hostname()
// ```
func Hostname() (name string, err error) {
	return os.Hostname()
}

// Unsetenv Delete the specified environment variable
// Example:
// ```
// os.Unsetenv("PATH")
// ```
func Unsetenv(key string) error {
	return os.Unsetenv(key)
}

// LookupEnv Get the specified The value of the environment variable
// Example:
// ```
// value, ok = os.LookupEnv("PATH")
// ```
func LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// Clearenv Clears all environment variables
// Example:
// ```
// os.Clearenv()
// ```
func Clearenv() {
	os.Clearenv()
}

// Setenv sets the specified environment variable
// Example:
// ```
// os.Setenv("PATH", "/usr/local/bin:/usr/bin:/bin")
// ```
func Setenv(key, value string) error {
	return os.Setenv(key, value)
}

// Getenv Gets the value of the specified environment variable, if not If it exists, return an empty string
// Example:
// ```
// value = os.Getenv("PATH")
// ```
func Getenv(key string) string {
	return os.Getenv(key)
}

// Exit Exit the current process
// Example:
// ```
// os.Exit(0)
// ```
func Exit(code int) {
	os.Exit(code)
}

// Args.
// Example:
// ```
// for arg in os.Args {
// println(arg)
// }
// ```
func osArgs() []string {
	return os.Args
}

// Executable Get the path of the current executable file
// Example:
// ```
// path, err = os.Executable()
// ```
func Executable() (string, error) {
	return os.Executable()
}

// ExpandEnv will string${var}or$var and replace it with the value of its corresponding environment variable name
// Example:
// ```
// os.ExpandEnv("PATH = $PATH")
// ```
func ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

// Pipe creates a pipe and returns a read Get end and a write end and error
// It is actually an alias of io.Pipe
// Example:
// ```
// r, w, err = os.Pipe()
// die(err)
//
//	go func {
//	    w.WriteString("hello yak")
//	    w.Close()
//	}
//
// bytes, err = io.ReadAll(r)
// die(err)
// dump(bytes)
// ```
func Pipe() (r *os.File, w *os.File, err error) {
	return os.Pipe()
}

// Chdir in Change the current working directory
// Example:
// ```
// err = os.Chdir("/tmp")
// ```
func Chdir(dir string) error {
	return os.Chdir(dir)
}

// Chmod Change the permissions of the specified file or directory
// Example:
// ```
// err = os.Chmod("/tmp/test.txt", 0777)
// ```
func Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}

// Chown changes the ownership of the specified file or directory User and group
// Example:
// ```
// err = os.Chown("/var/www/html/test.txt", 1000, 1000)
// ```
func Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

// GetDefaultDNSServers Get the string slice corresponding to the default DNS server IP
// Example:
// ```
// os.GetDefaultDNSServers()
// ```
func GetDefaultDNSServers() []string {
	return netx.NewDefaultReliableDNSConfig().SpecificDNSServers
}

// WaitConnect Wait for the port opening of an address or guide the timeout, and return an error if it times out. This is usually used to wait and ensure that a service starts
// Example:
// ```
// timeout, _ = time.ParseDuration("1m")
// ctx, cancel = context.WithTimeout(context.New(), timeout)
//
//	go func() {
//	    err = tcp.Serve("127.0.0.1", 8888, tcp.serverCallback(func (conn) {
//	    conn.Send("hello world")
//	    conn.Close()
//	}), tcp.serverContext(ctx))
//
//	    die(err)
//	}()
//
// os.WaitConnect("127.0.0.1:8888", 5)~ // Wait for the tcp server to start
// conn = tcp.Connect("127.0.0.1", 8888)~
// bytes = conn.Recv()~
// println(string(bytes))
// ```
func WaitConnect(addr string, timeout float64) error {
	return utils.WaitConnect(addr, timeout)
}

// Stdin Standard input
var Stdin = os.Stdin

// Stdout Standard output
var Stdout = os.Stdout

// Stderr standard error
var Stderr = os.Stderr

// OS The current operating system name
var OS = runtime.GOOS

// ARCH The running architecture of the current operating system: its value may be 386, amd64, arm, s390x, etc.
var ARCH = runtime.GOARCH

// IsPrivileged Whether the current mode is privileged
var IsPrivileged = privileged.GetIsPrivileged()

// GetLocalAddress Get the local IP address
// Example:
// ```
// os.GetLocalAddress() // ["192.168.1.103", "fe80::605a:5ff:fefb:5405"]
// ```
func GetLocalAddress() []string {
	ret, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	results := make([]string, len(ret))
	for i, a := range ret {
		if r, ok := a.(*net.IPNet); ok {
			results[i] = r.IP.String()
		}
	}
	return results
}

// GetLocalIPv4Address Get the local IPv4 Address
// Example:
// ```
// os.GetLocalIPv4Address() // ["192.168.3.103"]
// ```
func GetLocalIPv4Address() []string {
	var r []string
	for _, result := range GetLocalAddress() {
		if utils.IsLoopback(result) {
			continue
		}
		if utils.IsIPv4(result) {
			r = append(r, result)
		}
	}
	return r
}

// GetLocalIPv6Address Get the local IPv6 address
// Example:
// ```
// os.GetLocalIPv6Address() // ["fe80::605a:5ff:fefb:5405"]
// ```
func GetLocalIPv6Address() []string {
	var r []string
	for _, result := range GetLocalAddress() {
		if utils.IsLoopback(result) {
			continue
		}
		if utils.IsIPv6(result) {
			r = append(r, result)
		}
	}
	return r
}
