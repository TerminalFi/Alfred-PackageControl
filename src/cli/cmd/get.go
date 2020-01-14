package cmd

import (
	"os"
	"os/exec"
	"packagecontrol/packagecontrol"
	"strings"

	aw "github.com/deanishe/awgo"
	uuid "github.com/satori/go.uuid"
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

	client := packagecontrol.NewClient(nil)
	req, err := client.NewPackageRequest("GET", query)
	if err != nil {
		log.Info(err)
	}

	if err := client.Do(nil, req, &pkg); err != nil {
		log.Info(err)

	}

	if pkg.Name != "" {
		uuid4 := uuid.NewV4()
		wf.NewItem(pkg.Name).
			Subtitle(pkg.Description).
			Arg(pkg.Homepage).
			UID(uuid4.String()).
			Valid(true)

	}

	wf.WarnEmpty("No repos found", "Try a different package?")

	// Send results to Alfred
	wf.SendFeedback()

	return nil
}

func init() {
	rootCmd.AddCommand(getCmd)
}
