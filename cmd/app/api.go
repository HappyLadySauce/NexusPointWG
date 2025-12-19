package app

import (
	"fmt"
	"os"
	"context"

	"k8s.io/klog/v2"
	"github.com/spf13/cobra"
	"k8s.io/component-base/cli/flag"

	"github.com/HappyLadySauce/NexusPointWG/cmd/app/router"
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/options"
)

const (
	basename = "NexusPointWG"
)

func NewAPICommand(ctx context.Context) *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:   basename,
		Short: "NexusPointWG is a web server for WireGuard",
		Long:  "NexusPointWG is a web server for WireGuard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, opts)
		},
	}

	// 获取分组的 flags
	nfs := opts.Flags()

	// 将所有的 flags 添加到主 Command
	opts.AddFlags(cmd.Flags())

	// 设置 usage 和 help 函数以显示分组
	flag.SetUsageAndHelpFunc(cmd, *nfs, 80)

	return cmd
}

func run(ctx context.Context, opts *options.Options) error {

	<-ctx.Done()
	os.Exit(0)
	return nil
}

func serve(opts *options.Options) {
	insecureAddress := fmt.Sprintf("%s:%d", opts.InsecureServing.BindAddress, opts.InsecureServing.BindPort)
	klog.V(1).InfoS("Listening and serving on", "address", insecureAddress)
	go func() {
		klog.Fatal(router.Router().Run(insecureAddress))
	}()
}
