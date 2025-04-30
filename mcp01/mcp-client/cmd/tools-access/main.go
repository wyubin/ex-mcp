package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/wyubin/ex-mcp/mcp01/mcp-client/cmd/tools-access/cmd"
	"github.com/wyubin/ex-mcp/mcp01/utils/viperkit"
)

//go:embed env_default
var byteEnv []byte

func main() {
	viperkit.ReaderEnv(bytes.NewReader(byteEnv))
	viper.AutomaticEnv()
	err := cmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Println(err)
		os.Exit(99)
	}
}
