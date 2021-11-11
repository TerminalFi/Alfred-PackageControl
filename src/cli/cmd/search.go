package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"packagecontrol/packagecontrol"
	"strings"

	aw "github.com/deanishe/awgo"
	uuid "github.com/gofrs/uuid"
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

	if err := client.Do(context.TODO(), req, &packages); err != nil {
		log.Info(err)

	}

	for _, pkg := range packages.Packages {
		desc := fmt.Sprintf("%s installs", pkg.FormattedInstalls())
		if pkg.GetTrending() != 0 {
			desc = fmt.Sprintf("%s ⭐️ Trending", desc)
		}
		desc = fmt.Sprintf("%s  %s", desc, pkg.HighlightedDescription)

		uuid4 := uuid.Must(uuid.NewV4())
		item := wf.NewItem(pkg.GetName()).
			Subtitle(desc).
			Arg(fmt.Sprintf("https://packagecontrol.io/packages/%s", pkg.GetName())).
			UID(uuid4.String()).
			Valid(true)

		uuid4 = uuid.Must(uuid.NewV4())
		item.NewModifier(aw.ModCmd).
			Subtitle("Open Packages Homepage (Github, Gitlab, Bitbucket, Etc)").
			Arg(pkg.GetName()).
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
