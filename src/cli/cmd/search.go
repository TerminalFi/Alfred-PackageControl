package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"packagecontrol/packagecontrol"
	"strings"

	aw "github.com/deanishe/awgo"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search PackageControl.io for packages",
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: searchRun,
}

func searchRun(cobra *cobra.Command, args []string) error {
	query = strings.Join(args, " ")

	log.Printf("query=%s", query)

	// Call self with "check" command if an update is due and a check
	// job isn't already running.
	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")
		cmd := exec.Command(os.Args[0], "--update")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	// Only show update status if query is empty.
	if query == "" && wf.UpdateAvailable() {
		wf.Configure(aw.SuppressUIDs(true))

		wf.NewItem("Update available!").
			Subtitle("↩ to install").
			Autocomplete("workflow:update").
			Valid(false).
			Icon(iconUpdate)
	}

	var packages *packagecontrol.Packages
	client := packagecontrol.NewClient(nil)

	req, err := client.NewSearchRequest("GET", query)
	if err != nil {
		log.Info(err)
	}

	if err := client.Do(nil, req, &packages); err != nil {
		log.Info(err)

	}

	for _, pkg := range packages.Packages {
		uuid4 := uuid.NewV4()
		item := wf.NewItem(pkg.Name).
			Subtitle(pkg.HighlightedDescription).
			Arg(fmt.Sprintf("https://packagecontrol.io/packages/%s", pkg.Name)).
			UID(uuid4.String()).
			Valid(true)

		uuid4 = uuid.NewV4()
		item.NewModifier(aw.ModCmd).
			Subtitle("Open Packages Homepage (Github, Gitlab, Bitbucket, Etc)").
			Arg(pkg.Name).
			Valid(true)
	}
	wf.WarnEmpty("No repos found", "Try a different package?")

	// Send results to Alfred
	wf.SendFeedback()

	return nil
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
