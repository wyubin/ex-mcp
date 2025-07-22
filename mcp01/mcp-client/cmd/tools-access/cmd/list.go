package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wyubin/ex-mcp/mcp01/mcp-client/mcp"
	"github.com/wyubin/ex-mcp/mcp01/utils/log"
)

var (
	listProc *ListProc = NewListProc()
	listCmd            = &cobra.Command{
		Use:   "list [flags] pathConfig",
		Short: "list - list tools usage based servers in config",
		Long:  ``,
		Run:   listProc.run,
	}
)

type ListProc struct{}

func (s *ListProc) run(ccmd *cobra.Command, args []string) {
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
	codec := mcp.NewCfgCodec()
	var cfgServers mcp.CfgServers
	err = codec.Decode(byteConfig, &cfgServers)
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
	// init host
	host := mcp.NewHost()
	for nameServ, cfgServ := range cfgServers {
		err = host.SetClient(nameServ, cfgServ)
		if err != nil {
			log.Logger.Warn(err.Error())
		}
	}
	tools, err := host.ListTools()
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
	if len(tools) == 0 {
		log.Logger.Error("no tool exist")
		os.Exit(1)
	}
	byteTools, _ := json.Marshal(tools)
	fmt.Fprintf(os.Stdout, "%s\n", byteTools)
}

func NewListProc() *ListProc {
	return &ListProc{}
}
