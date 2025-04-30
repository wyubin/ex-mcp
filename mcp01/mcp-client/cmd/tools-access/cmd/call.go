package cmd

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
	"github.com/wyubin/ex-mcp/mcp01/mcp-client/core"
	"github.com/wyubin/ex-mcp/mcp01/utils/customflag"
	"github.com/wyubin/ex-mcp/mcp01/utils/log"
)

var (
	callProc *CallProc = NewCallProc()
	callCmd            = &cobra.Command{
		Use:   "call [flags] pathConfig",
		Short: "call - call tool with name and params based servers in config",
		Long:  ``,
		Run:   callProc.run,
	}
	nameTool string
	params   customflag.FlagJsonMap
)

type CallProc struct{}

func (s *CallProc) run(ccmd *cobra.Command, args []string) {
	// use first arg as server config
	if len(args) < 1 {
		log.Logger.Error("need one config path")
		os.Exit(1)
	}
	byteConfig, err := os.ReadFile(args[0])
	if err != nil {
		log.Logger.Error(fmt.Sprintf("config not accessible: %s", err.Error()))
		os.Exit(1)
	}
	codec := core.NewCfgCodec()
	var cfgServers core.CfgServers
	err = codec.Decode(byteConfig, &cfgServers)
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
	// init host
	host := core.NewHost()
	for nameServ, cfgServ := range cfgServers {
		err = host.SetClient(nameServ, cfgServ)
		if err != nil {
			log.Logger.Warn(err.Error())
		}
	}
	// call tool
	rawContents, err := host.CallTool(nameTool, params)
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
	content := rawContents[0].(mcp.TextContent)
	fmt.Fprintf(os.Stdout, "%s\n", content.Text)
}

func NewCallProc() *CallProc {
	return &CallProc{}
}

func init() {
	persistFlag := callCmd.PersistentFlags()
	persistFlag.StringVar(&nameTool, "name-tool", "<server>.<tool>", "assign tool name to use")
	callCmd.MarkPersistentFlagRequired("name-tool")
	persistFlag.Var(&params, "params", "define parameters for tool with json format")
	callCmd.MarkPersistentFlagRequired("params")
}
