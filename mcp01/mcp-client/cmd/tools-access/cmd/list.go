package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wyubin/ex-mcp/mcp01/mcp-client/core"
	"github.com/wyubin/ex-mcp/mcp01/utils/log"
)

var (
	listProc *ListProc = NewListAdd()
	listCmd            = &cobra.Command{
		Use:   "add-sample [flags] pathVcf [...pathVcf]",
		Short: "add sample - AF data from VCF file",
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
	tools, err := host.ListTools()
	if err != nil {
		log.Logger.Error(err.Error())
		os.Exit(1)
	}
	if len(tools) == 0 {
		log.Logger.Error("no tool exist")
		os.Exit(1)
	}
	fmt.Printf("tools: %+v\n", tools)
}

func NewListAdd() *ListProc {
	return &ListProc{}
}
