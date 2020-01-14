package cmd

import (
	"fmt"
	"os"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const updateJobName = "checkForUpdate"

var (
	wf *aw.Workflow

	checkUpdate bool
	query       string

	repo       = "thesecenv/alfred-packagecontrol" // GitHub repo
	iconUpdate = &aw.Icon{Value: "update-available.png"}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "packagecontrol",
	Short: "Search the Sublime PackageControl.io Website",
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		checkUpdate, err := cmd.Flags().GetBool("update")
		if err != nil {
			wf.FatalError(err)
		}
		// Alternate action: Get available releases from remote.
		if checkUpdate {
			wf.Configure(aw.TextErrors(true))
			log.Println("Checking for updates...")
			if err := wf.CheckForUpdate(); err != nil {
				wf.FatalError(err)
			}
			return
		}
	},
}

func Execute() {
	wf.Run(func() {
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})
}

func init() {
	wf = aw.New(
		update.GitHub(repo),
	)

	wf.Args()

	rootCmd.Flags().BoolP("update", "u", false, "Check for workflow updates")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
