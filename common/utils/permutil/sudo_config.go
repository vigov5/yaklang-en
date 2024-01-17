package permutil

import (
	"context"
	"io"
	"os"
)

type SudoConfig struct {
	// Prompt, generally used to tell the user what this Sudo is used for.
	Verbose string

	// CWD: What is the directory where the command is executed?
	Workdir string

	// Env
	Environments map[string]string

	// is used to control the life cycle.
	Ctx context.Context

	Stdout, Stderr  io.Writer
	ExitCodeHandler func(int)
}

type SudoOption func(config *SudoConfig)

func WithVerbose(i string) SudoOption {
	return func(config *SudoConfig) {
		config.Verbose = i
	}
}

func WithWorkdir(i string) SudoOption {
	return func(config *SudoConfig) {
		config.Workdir = i
	}
}

func WithContext(ctx context.Context) SudoOption {
	return func(config *SudoConfig) {
		config.Ctx = ctx
	}
}

func WithEnv(k, v string) SudoOption {
	return func(config *SudoConfig) {
		if config.Environments == nil {
			config.Environments = make(map[string]string)
		}
		config.Environments[k] = v
	}
}

func NewDefaultSudoConfig() *SudoConfig {
	return &SudoConfig{
		Verbose:      "Auth(or Password) Required",
		Workdir:      os.TempDir(),
		Environments: make(map[string]string),
		Ctx:          context.Background(),
	}
}

func WithStdout(w io.Writer) SudoOption {
	return func(config *SudoConfig) {
		config.Stdout = w
	}
}

func WithStderr(w io.Writer) SudoOption {
	return func(config *SudoConfig) {
		config.Stderr = w
	}
}

func WithExitCodeHandler(i func(int)) SudoOption {
	return func(config *SudoConfig) {
		config.ExitCodeHandler = i
	}
}
