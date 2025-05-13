package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	codecDefault = NewCfgCodec()
)

func TestCfgCodec(t *testing.T) {
	cfgServers := CfgServers{
		"sseTest": cfgSSE,
	}
	byteConfig, err := codecDefault.Encode(&cfgServers)
	assert.NoError(t, err, "TestCfgCodec - Encode")
	fmt.Printf("byteConfig: %s\n", byteConfig)

	cfgTmp := &CfgServers{}
	err = codecDefault.Decode(byteConfig, cfgTmp)
	assert.NoError(t, err, "TestCfgCodec - Decode")
	fmt.Printf("cfgTmp: %+v\n", cfgTmp)
}
