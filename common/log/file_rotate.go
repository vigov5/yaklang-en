package log

import (
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"io"
	"os"
	"time"
)

func InitFileRotateLogger(baseLogPath string, logSaveDay int, logRotateHour int) error {
	writer, err := rotatelogs.New(
		baseLogPath+".%Y-%m-%d_%H_%M",
		rotatelogs.WithLinkName(baseLogPath),                                // Generate soft link pointing to the latest log file
		rotatelogs.WithMaxAge(time.Duration(logSaveDay)*24*time.Hour),       // Maximum file storage time
		rotatelogs.WithRotationTime(time.Duration(logRotateHour)*time.Hour), // Log cutting time interval
	)
	if err != nil {
		return errors.Errorf("init file rotate logger failed: %s", err)
	}

	w := io.MultiWriter(writer, os.Stdout)
	SetOutput(w)
	return nil
}
