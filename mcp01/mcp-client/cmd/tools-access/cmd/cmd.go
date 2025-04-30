package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/wyubin/ex-mcp/mcp01/utils/log"
)

const (
	NAME_CMD = "mcptools"
)

var (
	ckDebug bool

	ctlCmd = &cobra.Command{
		Use:           NAME_CMD,
		Short:         fmt.Sprintf("%s – tools list/execute of MCP", NAME_CMD),
		Long:          `based mcp servers config to get tools or run tool`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	defer log.LogExeTime(NAME_CMD)()
	return ctlCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	// add subcmd
	ctlCmd.AddCommand(listCmd)
	ctlCmd.AddCommand(callCmd)
}

func initConfig() {
	// init logger
	var logLevel slog.Level = slog.LevelInfo
	if ckDebug {
		logLevel = slog.LevelDebug
	}
	log.InitLogger(logLevel, os.Stderr)
}
