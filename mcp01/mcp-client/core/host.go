package core

// 實作 host 控制每個 client 的操作，也會將不同 client 的工具描述再一次輸出給外部
type Host struct{}

// add single server with name
// 如果有重複名字的server 就會 overwrite
func (s *Host) SetClient(name string, cfgServ CfgServer) error {
	return nil
}

// Get Client, 如果是 nil 代表不存在
func (s *Host) GetClient(name string) *Client {
	return nil
}

// Get Clients
