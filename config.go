package socks5

type SockInfo struct {
	Addr     string `json:"addr,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}
type Filter struct {
	Include []struct {
		Type  string `json:"type,omitempty"` // regexp cidr
		Value string `json:"value,omitempty"`
	} `json:"include,omitempty"`
	Exclude []struct {
		Type  string `json:"type,omitempty"` // regexp cidr
		Value string `json:"value,omitempty"`
	} `json:"exclude,omitempty"`
}
type Socks struct {
	SockInfo
	Filter         // TODO 如果父级为空，将子集提到父级而不用多配置几次。
	Chains []Socks `json:"chains,omitempty"` // todo 链式访问服务
}
type ConfigInfo struct {
	Addr  string  `json:"addr,omitempty"`
	Socks []Socks `json:"socks"`
}
