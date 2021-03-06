package cmd

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm"
	pfscmds "github.com/pachyderm/pachyderm/src/pfs/cmds"
	deploycmds "github.com/pachyderm/pachyderm/src/pkg/deploy/cmds"
	ppscmds "github.com/pachyderm/pachyderm/src/pps/cmds"
	"github.com/spf13/cobra"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/pkg/cobra"
	"go.pedge.io/proto/version"
	"golang.org/x/net/context"
)

func PachctlCmd(address string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use: os.Args[0],
		Long: `Access the Pachyderm API.

Envronment variables:
  ADDRESS=0.0.0.0:650, the server to connect to.
`,
	}
	pfsCmds := pfscmds.Cmds(address)
	for _, cmd := range pfsCmds {
		rootCmd.AddCommand(cmd)
	}
	ppsCmds, err := ppscmds.Cmds(address)
	if err != nil {
		return nil, err
	}
	for _, cmd := range ppsCmds {
		rootCmd.AddCommand(cmd)
	}

	deployCmds := deploycmds.Cmds()
	for _, cmd := range deployCmds {
		rootCmd.AddCommand(cmd)
	}
	version := &cobra.Command{
		Use:   "version",
		Short: "Return version information.",
		Long:  "Return version information.",
		Run: pkgcobra.RunFixedArgs(0, func(args []string) error {
			versionClient, err := getVersionAPIClient(address)
			if err != nil {
				return err
			}
			version, err := versionClient.GetVersion(context.Background(), &google_protobuf.Empty{})
			if err != nil {
				return err
			}
			writer := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
			printVerisonHeader(writer)
			printVersion(writer, "pachctl", pachyderm.Version)
			printVersion(writer, "pachd", version)
			return writer.Flush()
		}),
	}
	rootCmd.AddCommand(version)
	return rootCmd, nil
}

func getVersionAPIClient(address string) (protoversion.APIClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return protoversion.NewAPIClient(clientConn), nil
}

func printVerisonHeader(w io.Writer) {
	fmt.Fprintf(w, "COMPONENT\tVERSION\t\n")
}

func printVersion(w io.Writer, component string, version *protoversion.Version) {
	fmt.Fprintf(
		w,
		"%s\t%d.%d.%d(%s)\t\n",
		component,
		version.Major,
		version.Minor,
		version.Micro,
		version.Additional,
	)
}
