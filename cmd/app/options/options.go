package options

import (
	"encoding/json"

	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"

	"github.com/HappyLadySauce/NexusPointWG/pkg/options"
)

type Options struct {
	InsecureServing *options.InsecureServingOptions

}

func NewOptions() *Options {
	return &Options{
		InsecureServing: options.NewInsecureServingOptions(),
	}
}

// Flags 返回 NamedFlagSets，用于分组显示 flags
func (o *Options) Flags() *flag.NamedFlagSets {
	nfs := &flag.NamedFlagSets{}

	// 创建标志分组并添加配置标志
	configFS := nfs.FlagSet("Config")
	options.AddConfigFlag(configFS)

	insecureServingFS := nfs.FlagSet("Insecure Serving")
	o.InsecureServing.AddFlags(insecureServingFS)

	return nfs
}

// AddFlags 将所有的 flags 添加到指定的 FlagSet 中（用于向后兼容）
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	nfs := o.Flags()
	// 将 NamedFlagSets 中的所有 flags 添加到主 FlagSet
	for _, name := range nfs.Order {
		fs.AddFlagSet(nfs.FlagSets[name])
	}
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)

	return string(data)
}
