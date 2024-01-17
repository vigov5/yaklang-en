package yaktest

import "testing"

func TestPacketBatchScan(t *testing.T) {
	var test = `yakit.AutoInitYakit()

loglevel("info")
productMode = cli.Bool("prod")

// Get flowsID
flows = make([]int64)
flowsFile = cli.String("flows-file")
if flowsFile != "" {
    ports = str.ParseStringToPorts(string(file.ReadFile(flowsFile)[0]))
    for _, port = range ports {
        flows = append(flows, int64(port))
    }
}

if !productMode { yakit.StatusCard("Debug mode", "TRUE") }

if len(flows) <= 0 && !productMode {
    yakit.Info("Debug mode Supplement FlowID")
    flows = append(flows, 17)
}

caller, err = hook.NewMixPluginCaller()
die(err)

// Load plug-in
plugins = cli.YakitPlugin() // yakit-plugin-file
if len(plugins) <= 0 && !productMode {
    // Set the default plug-in, generally used here for debugging
    plugins = [
        "WebLogic_CVE-2022-21371 directory traversal vulnerability detection",
    ]
}
loadedPlugins = []
failedPlugins = []
for _, pluginName = range x.If(plugins == undefined, [], plugins) {
    err := caller.LoadPlugin(pluginName)
    if err != nil {
        failedPlugins = append(failedPlugins, pluginName)
        yakit.Error("load plugin[%s] failed: %s", pluginName, err)
    }else{
        yakit.Info("Load plug-in: %v", pluginName)
        loadedPlugins = append(loadedPlugins, pluginName)
    }
}

if len(failedPlugins)>0 {
    yakit.StatusCard("Number of plug-in loading failures", len(failedPlugins))
}

if len(loadedPlugins) <= 0 {
    yakit.StatusCard("Execution failed", "No plug-in is loaded Correctly load", "error")
    die("No plug-in is loaded")
}else{
    yakit.StatusCard("Number of plug-in loads", len(loadedPlugins))
}

packetRawFile = cli.String("packet-file")
packetHttps = cli.Bool("packet-https")

// Set concurrency parameters
pluginConcurrent = 20
packetConcurrent = 20

// Set asynchronous, set concurrency, set Wait
caller.SetDividedContext(true)
caller.SetConcurrent(pluginConcurrent)
defer caller.Wait()

handleHTTPFlow = func(flow) {
    defer func{
        err = recover()
        if err != nil {
            yakit.Error("handle httpflow failed: %s", err)
        }
    }

    reqBytes = []byte(codec.StrconvUnquote(flow.Request)[0])
    rspBytes = []byte(codec.StrconvUnquote(flow.Response)[0])
    _, body = poc.Split(rspBytes)
    // Call the mirror traffic function and control whether to perform port scanning The plug-in calls
    caller.MirrorHTTPFlowEx(true, flow.IsHTTPS, flow.Url, reqBytes, rspBytes, body)
}

handleHTTPRequestRaw = func(https, reqBytes) {
    defer func{
        err = recover()
        if err != nil {
            yakit.Error("handle raw bytes failed: %s", err)
        }
    }

    urlstr = ""
    urlIns, _ = str.ExtractURLFromHTTPRequestRaw(reqBytes, https)
    if urlIns != nil {
        urlstr = urlIns.String()
    }
    rsp, req, _ = poc.HTTP(reqBytes, poc.https(packetHttps))
    header, body = poc.Split(rsp)
    caller.MirrorHTTPFlowEx(true, https, urlstr, reqBytes, rsp, body)
}

// Start getting HTTPFlow ID
if len(flows) > 0 {
    yakit.StatusCard("Scan historical data packets", len(flows))
	for result = range db.QueryHTTPFlowsByID(flows...) {
		// Submit scanning task:
		yakit.Info("Submit scanning task: [%v] %v", result.Method, result.Url)
		handleHTTPFlow(result)
	}
}

if packetRawFile != "" {
    raw, err = file.ReadFile(packetRawFile)
    if err != nil {
        yakit.StatusCard("Reason for failure", "cannot receive parameters", "error")
    }
    handleHTTPRequestRaw(packetHttps, raw)
}
`
	Run("Test packet test", t, []YakTestCase{
		{
			Name: "Test packet test",
			Src:  test,
		},
	}...)
}
