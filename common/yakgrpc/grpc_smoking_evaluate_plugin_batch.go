package yakgrpc

import (
	"encoding/json"
	"fmt"

	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

// SmokingEvaluatePluginBatch(*SmokingEvaluatePluginBatchRequest, Yak_SmokingEvaluatePluginBatchServer) error
// SmokingEvaluatePluginBatch
func (s *Server) SmokingEvaluatePluginBatch(req *ypb.SmokingEvaluatePluginBatchRequest, stream ypb.Yak_SmokingEvaluatePluginBatchServer) error {
	// fmt.Println("in smoking evaluate plugin batch!!")
	send := func(progress float64, message, messageType string) {
		// fmt.Println("progress: ", progress, " message: ", message, " messageType: ", messageType)
		stream.Send(&ypb.SmokingEvaluatePluginBatchResponse{
			Progress:    progress,
			Message:     message,
			MessageType: messageType,
		})
	}
	names := make([]string, 0, len(req.GetScriptNames()))
	successNum := 0
	errorNum := 0

	send(0, "started detection", "success")
	pluginSize := len(req.GetScriptNames())
	for index, name := range req.GetScriptNames() {
		progress := float64(index+1) / float64(pluginSize)
		// fmt.Println("check name: ", index, name, pluginSize, progress)
		ins, err := yakit.GetYakScriptByName(s.GetProfileDatabase(), name)
		if err != nil {
			msg := fmt.Sprintf("%s: Unable to obtain the plug-in", name)
			send(progress, msg, "error")
			errorNum++
			continue
		}
		code := ins.Content
		pluginType := ins.Type
		res, err := s.EvaluatePlugin(stream.Context(), code, pluginType)
		if err != nil {
			msg := fmt.Sprintf("%s failed to start plug-in detection", name)
			send(progress, msg, "error")
			errorNum++
			continue
		}
		if res.Score >= 60 {
			msg := fmt.Sprintf("%s Plug-in score: %d", name, res.Score)
			send(progress, msg, "success")
			names = append(names, name)
			successNum++
			continue
		} else {
			msg := fmt.Sprintf("%s Plug-in score: %d (<60)", name, res.Score)
			send(progress, msg, "error")
			errorNum++
			continue
		}
	}

	{
		msg := ""
		if successNum > 0 {
			msg += fmt.Sprintf("%d detection passed", successNum)
		}
		if errorNum > 0 {
			msg += fmt.Sprintf(", %d detection failed", errorNum)
		}
		if msg == "" {
			msg += "detection ended"
		}
		send(1, msg, "success")
	}
	msg, err := json.Marshal(names)
	if err != nil {
		return err
	}
	send(2, utils.UnsafeBytesToString(msg), "success-again")
	// send(2, strings.Join(names, ","), "success-again")
	return nil
}
