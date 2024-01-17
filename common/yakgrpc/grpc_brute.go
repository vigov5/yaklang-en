package yakgrpc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/bruteutils"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

const startBruteScript = `yakit.AutoInitYakit()

debug = false

yakit.Info("Start checking execution parameters")

targetFile := cli.String("target-file", cli.setRequired(true))
userList := cli.String("user-list-file")
passList := cli.String("pass-list-file")
concurrent := cli.Int("concurrent")
taskConcurrent := cli.Int("task-concurrent")
minDelay, maxDelay := cli.Int("delay-min", cli.setDefault(3)), cli.Int("delay-max", cli.setDefault(5))
pluginFile := cli.String("plugin-file")
okToStop := cli.Bool("ok-to-stop")
replaceDefaultUsernameDict := cli.Bool("replace-default-username-dict")
replaceDefaultPasswordDict := cli.Bool("replace-default-password-dict")
finishingThreshold = cli.Int("finishing-threshold", cli.setDefault(1))

yakit.Info("Check blasting Type")
bruteTypes = cli.String("types")
if bruteTypes == "" {
    yakit.Error("No Explosion Type specified")
    if !debug {
        die("exit normal")
    }
    bruteTypes = "ssh"
}

// TargetsConcurrent
// TargetTaskConcurrent
// DelayerMin DelayerMax
// BruteCallback
// OkToStop
// FinishingThreshold
// OnlyNeedPassword
wg := sync.NewWaitGroup()
defer wg.Wait()

yakit.Info("Scan target preprocessing")
// Process the scan target
raw, _ := file.ReadFile(targetFile)
if len(raw) == 0 {
    yakit.Error("BUG: Failed to read target file!")
    if !debug {
        return
    }
    raw = []byte("127.0.0.1:23")
}
target = str.ParseStringToLines(string(raw))

targetRaw = make([]string)
for _, t := range target {
	if t.Contains("://") {
		targetRaw = append(targetRaw, t)
	} else {
		host, port, err := str.ParseStringToHostPort(t)
		if err != nil {
			targetRaw = append(targetRaw, t)
		}else{
			targetRaw = append(targetRaw, str.HostPort(host, port))
		}
	}
}
target = targetRaw

yakit.Info("User-defined dictionary preprocessing")
// Define the dictionary that stores usernames and passwords
userdefinedUsernameList = make([]string)
userdefinedPasswordList = make([]string)

// Get the user list
userRaw, _ := file.ReadFile(userList)
if len(userRaw) <= 0 {
    yakit.Error("User file dictionary acquisition failed")
}else{
    userdefinedUsernameList = str.ParseStringToLines(string(userRaw))
}

// Obtain user password
passRaw, _ := file.ReadFile(passList)
if len(passRaw) <= 0 {
    yakit.Error("User password file acquisition failed")
}else{
    userdefinedPasswordList = str.ParseStringToLines(string(passRaw))
}

opt = []

if minDelay > 0 && maxDelay > 0 {
    yakit.Info("Single target test random delay: %v-%v/s", minDelay, maxDelay)
    opt = append(opt, brute.minDelay(minDelay), brute.maxDelay(maxDelay))
}

if finishingThreshold > 0 {
    opt = append(opt, brute.finishingThreshold(finishingThreshold))
}

if concurrent > 0 {
    yakit.Info("Set the maximum number of simultaneous blasting targets: %v", concurrent)
    opt = append(opt, brute.concurrentTarget(concurrent))
}

if taskConcurrent > 0 {
    yakit.Info("Set single target blasting concurrency: %v", taskConcurrent)
    opt = append(opt, brute.concurrent(taskConcurrent))
}


tableName = "Available blasting result table"
columnType = "TYPE"
columnTarget = "TARGET"
columnUsername = "USERNAME"
columnPassword = "PASSWORD"
yakit.EnableTable(tableName, [columnType, columnTarget, columnUsername, columnPassword])

scan = func(bruteType) {
    yakit.Info("Enable Exploit Program for %v", bruteType)
    wg.Add(1)
    go func{
        defer wg.Done()

        tryCount = 0
        success = 0
        failed = 0
        finished = 0

        uL = make([]string)
        pL = make([]string)
        if (!replaceDefaultUsernameDict) {
            uL = append(uL, brute.GetUsernameListFromBruteType(bruteType)...)
        }

        if (!replaceDefaultPasswordDict) {
			pL = append(pL, brute.GetPasswordListFromBruteType(bruteType)...)
        }

        instance, err := brute.New(
            string(bruteType),
            brute.userList(append(userdefinedUsernameList, uL...)...),
            brute.passList(append(userdefinedPasswordList, pL...)...),
            brute.debug(true),
            brute.okToStop(okToStop),
            opt...
        )
        if err != nil {
            yakit.Error("Failed to construct weak passwords and unauthorized scanning: %v", err)
            return
        }

        res, err := instance.Start(target...)
        if err != nil {
            yakit.Error("Enter target failed: %v", err)
            return
        }

        for result := range res {
            tryCount++
            yakit.StatusCard("Total attempts: "+bruteType, tryCount, bruteType, "total")
            result.Show()

            if result.Ok {
                success++
                yakit.StatusCard("Number of successes: "+bruteType, success, bruteType, "success")
				if result.Username == "" && result.Password == "" {
					risk.NewRisk(
						result.Target, risk.severity("high"), risk.type("weak-pass"),
						risk.typeVerbose("Unauthorized access"),
						risk.title(sprintf("Unauthorized access [%v]: %v", result.Type, result.Target)),
						risk.titleVerbose(sprintf("Unauthorized access [%v]: %v", result.Type, result.Target)),
						risk.description("Due to improper configuration or management negligence, there is a risk of unauthorized access to certain services, interfaces or applications. Attackers can directly access these resources without any authentication, which may lead to the disclosure of sensitive data, abuse of the system, or other malicious behavior."),
risk.solution(` + "`" + `1. Audit all publicly accessible services, interfaces, and applications to ensure they have appropriate access controls.
2. Use authentication mechanism, such as username/Password, API key or OAuth.
3. Regularly monitor and review access logs to detect any suspicious or unauthorized activity. ` + "`" + `),
						risk.details({"target": result.Target}),
					)
				} else {
					risk.NewRisk(
						result.Target, risk.severity("high"), risk.type("weak-pass"),
						risk.typeVerbose("Weak password"),
						risk.title(sprintf("Weak Password[%v]：%v user(%v) pass(%v)", result.Type, result.Target, result.Username, result.Password)),
						risk.titleVerbose(sprintf("weak password [%v]: %v user(%v) pass(%v)", result.Type, result.Target, result.Username, result.Password)),
						risk.details({"username": result.Username, "password": result.Password, "target": result.Target}),
					)
				}
               
                yakit.Output(yakit.TableData(tableName, {
                    columnType: result.Type,
                    columnTarget: result.Target,
                    columnUsername: result.Username,
                    columnPassword: result.Password,
                    "id": tryCount,
                    "bruteType": bruteType,
                }))
            } else {
                failed++
                yakit.StatusCard("Number of failures: " + bruteType, failed, bruteType, "failed")
            }
        }
    }
}

for _, t := range str.Split(bruteTypes, ",") {
    scan(t)
}`

func (s *Server) StartBrute(params *ypb.StartBruteParams, stream ypb.Yak_StartBruteServer) error {
	reqParams := &ypb.ExecRequest{Script: startBruteScript}

	types := utils.PrettifyListFromStringSplited(params.GetType(), ",")
	for _, t := range types {
		h, err := bruteutils.GetBruteFuncByType(t)
		if err != nil || h == nil {
			return utils.Errorf("brute type: %v is not available", t)
		}
	}
	reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "types", Value: params.GetType()})

	targetFile, err := utils.DumpHostFileWithTextAndFiles(params.Targets, "\n", params.TargetFile)
	if err != nil {
		return err
	}
	defer os.RemoveAll(targetFile)
	reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "target-file", Value: targetFile})

	// resolves the user name.
	userListFile, err := utils.DumpFileWithTextAndFiles(
		strings.Join(params.Usernames, "\n"), "\n", params.UsernameFile,
	)
	if err != nil {
		return err
	}
	defer os.RemoveAll(userListFile)
	reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "user-list-file", Value: userListFile})

	// use the default dictionary?
	if params.GetReplaceDefaultPasswordDict() {
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "replace-default-password-dict"})
	}

	if params.GetReplaceDefaultUsernameDict() {
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "replace-default-username-dict"})
	}

	// parses passwords.
	passListFile, err := utils.DumpFileWithTextAndFiles(
		strings.Join(params.Passwords, "\n"), "\n", params.PasswordFile,
	)
	if err != nil {
		return err
	}
	defer os.RemoveAll(passListFile)
	reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "pass-list-file", Value: passListFile})

	// ok to stop
	if params.GetOkToStop() {
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "ok-to-stop", Value: ""})
	}

	if params.GetConcurrent() > 0 {
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "concurrent", Value: fmt.Sprint(params.GetConcurrent())})
	}

	if params.GetTargetTaskConcurrent() > 0 {
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "task-concurrent", Value: fmt.Sprint(params.GetTargetTaskConcurrent())})
	}

	if params.GetDelayMin() > 0 && params.GetDelayMax() > 0 {
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "delay-min", Value: fmt.Sprint(params.GetDelayMin())})
		reqParams.Params = append(reqParams.Params, &ypb.ExecParamItem{Key: "delay-max", Value: fmt.Sprint(params.GetDelayMax())})
	}

	return s.Exec(reqParams, stream)
}

func (s *Server) GetAvailableBruteTypes(ctx context.Context, req *ypb.Empty) (*ypb.GetAvailableBruteTypesResponse, error) {
	return &ypb.GetAvailableBruteTypesResponse{Types: bruteutils.GetBuildinAvailableBruteType()}, nil
}
