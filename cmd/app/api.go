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

	nfs := opts.AddFlags(cmd.Flags())

	flag.SetUsageAndHelpFunc(cmd, *nfs, 80)

	return cmd
}

func run(ctx context.Context, opts *options.Options) error {
	serve(opts)
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
