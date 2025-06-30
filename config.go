package socks5

type Socks struct {
	Addr     string `json:"addr,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Network  string `json:"network,omitempty"`
	Include  []struct {
		Type  string `json:"type,omitempty"` // regexp cidr
		Value string `json:"value,omitempty"`
	} `json:"include,omitempty"`
	Exclude []struct {
		Type  string `json:"type,omitempty"` // regexp cidr
		Value string `json:"value,omitempty"`
	} `json:"exclude,omitempty"`
}
type ConfigInfo struct {
	Addr  string  `json:"addr,omitempty"`
	Socks []Socks `json:"socks"`
}
