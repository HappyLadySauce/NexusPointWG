package options

import (
	"fmt"
	"net"
	"github.com/spf13/pflag"
)

type InsecureServingOptions struct {
	BindAddress net.IP `json:"bind-address" mapstructure:"bind-address"`
	BindPort    int    `json:"bind-port"    mapstructure:"bind-port"`
}

func NewInsecureServingOptions() *InsecureServingOptions {
	return &InsecureServingOptions{}
}


func (i *InsecureServingOptions) Validate() []error {
	var errors []error
	if i.BindAddress == nil {
		errors = append(errors, fmt.Errorf("bind-address is required"))
	}
	if i.BindPort == 0 {
		errors = append(errors, fmt.Errorf("bind-port is required"))
	}
	return errors
}	

func (i *InsecureServingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IPVarP(&i.BindAddress, "bind-address", "b", net.ParseIP("127.0.0.1"), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	fs.IntVarP(&i.BindPort, "bind-port", "p", 8001, "port to listen to for incoming HTTPS requests")
}