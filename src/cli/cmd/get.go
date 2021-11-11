package cmd

import (
	"fmt"
	"packagecontrol/packagecontrol"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve a packages Homepage URL",
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: getRun,
}

func getRun(cobra *cobra.Command, args []string) error {
	var pkg packagecontrol.PackageDetails

	query := strings.Join(args, " ")

	log.Printf("query=%s", query)

	client := packagecontrol.NewClient(nil)
	req, err := client.NewPackageRequest("GET", query)
	if err != nil {
		return err
	}

	if err := client.Do(nil, req, &pkg); err != nil {
		return err
	}

	if pkg.Name != "" {
		fmt.Print(pkg.GetURL())
	}
	return nil
}

func init() {
	rootCmd.AddCommand(getCmd)
}
