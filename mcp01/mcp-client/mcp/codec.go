package mcp

import (
	"encoding/json"
)

const (
	KEY_CONFIG_SERVERS = "mcpServers"
)

// 實作 encode 跟decode
type CfgCodec struct {
	key string
}

func NewCfgCodec(keys ...string) *CfgCodec {
	key := KEY_CONFIG_SERVERS
	if len(keys) > 0 {
		key = keys[0]
	}
	inst := CfgCodec{key: key}
	return &inst
}

func (s *CfgCodec) Encode(cfgPt interface{}) ([]byte, error) {
	cfgs, ok := cfgPt.(*CfgServers)
	if !ok {
		return nil, ErrInValidCfgServers
	}
	config := map[string]*CfgServers{s.key: cfgs}
	return json.Marshal(config)
}

func (s *CfgCodec) Decode(byteConfig []byte, cfgPt interface{}) error {
	var config map[string]interface{}
	err := json.Unmarshal(byteConfig, &config)
	if err != nil {
		return ErrInValidCfgServers
	}
	cfgServersRaw, ok := config[s.key]
	if !ok {
		return ErrInValidCfgServers
	}
	byteServers, _ := json.Marshal(cfgServersRaw)

	return json.Unmarshal(byteServers, cfgPt)
}
