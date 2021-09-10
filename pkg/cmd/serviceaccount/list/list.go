package list

import (
	"context"
	"encoding/json"

	"github.com/redhat-developer/app-services-cli/internal/config"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/factory"
	"github.com/redhat-developer/app-services-cli/pkg/cmd/flag"
	"github.com/redhat-developer/app-services-cli/pkg/cmdutil"
	flagutil "github.com/redhat-developer/app-services-cli/pkg/cmdutil/flags"
	"github.com/redhat-developer/app-services-cli/pkg/connection"
	"github.com/redhat-developer/app-services-cli/pkg/dump"
	"github.com/redhat-developer/app-services-cli/pkg/iostreams"
	"github.com/redhat-developer/app-services-cli/pkg/localize"
	"github.com/redhat-developer/app-services-cli/pkg/logging"
	kafkamgmtclient "github.com/redhat-developer/app-services-sdk-go/kafkamgmt/apiv1/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type options struct {
	Config     config.IConfig
	Connection factory.ConnectionFunc
	Logger     logging.Logger
	IO         *iostreams.IOStreams
	localizer  localize.Localizer
	Context    context.Context

	output string
}

// svcAcctRow contains the properties used to
// populate the list of service accounts into a table row
type svcAcctRow struct {
	ID        string `json:"id" header:"ID"`
	Name      string `json:"name" header:"Name"`
	ClientID  string `json:"clientID" header:"Client ID"`
	Owner     string `json:"owner" header:"Owner"`
	CreatedAt string `json:"createdAt" header:"Created At"`
}

// NewListCommand creates a new command to list service accounts
func NewListCommand(f *factory.Factory) *cobra.Command {
	opts := &options{
		Config:     f.Config,
		Connection: f.Connection,
		Logger:     f.Logger,
		IO:         f.IOStreams,
		localizer:  f.Localizer,
		Context:    f.Context,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   opts.localizer.MustLocalize("serviceAccount.list.cmd.shortDescription"),
		Long:    opts.localizer.MustLocalize("serviceAccount.list.cmd.longDescription"),
		Example: opts.localizer.MustLocalize("serviceAccount.list.cmd.example"),
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.output != "" && !flagutil.IsValidInput(opts.output, flagutil.ValidOutputFormats...) {
				return flag.InvalidValueError("output", opts.output, flagutil.ValidOutputFormats...)
			}

			return runList(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.output, "output", "o", "", opts.localizer.MustLocalize("serviceAccount.list.flag.output.description"))

	flagutil.EnableOutputFlagCompletion(cmd)

	return cmd
}

func runList(opts *options) (err error) {
	conn, err := opts.Connection(connection.DefaultConfigSkipMasAuth)
	if err != nil {
		return err
	}

	res, _, err := conn.API().ServiceAccount().GetServiceAccounts(opts.Context).Execute()
	if err != nil {
		return err
	}

	serviceaccounts := res.GetItems()
	if len(serviceaccounts) == 0 && opts.output == "" {
		opts.Logger.Info(opts.localizer.MustLocalize("serviceAccount.list.log.info.noneFound"))
		return nil
	}

	outStream := opts.IO.Out
	switch opts.output {
	case dump.JSONFormat:
		data, _ := json.MarshalIndent(res, "", cmdutil.DefaultJSONIndent)
		_ = dump.JSON(outStream, data)
	case dump.YAMLFormat, dump.YMLFormat:
		data, _ := yaml.Marshal(res)
		_ = dump.YAML(outStream, data)
	default:
		rows := mapResponseItemsToRows(serviceaccounts)
		dump.Table(outStream, rows)
	}

	return nil
}

func mapResponseItemsToRows(svcAccts []kafkamgmtclient.ServiceAccountListItem) []svcAcctRow {
	rows := []svcAcctRow{}

	for _, sa := range svcAccts {
		row := svcAcctRow{
			ID:        sa.GetId(),
			Name:      sa.GetName(),
			ClientID:  sa.GetClientId(),
			Owner:     sa.GetOwner(),
			CreatedAt: sa.GetCreatedAt().String(),
		}

		rows = append(rows, row)
	}

	return rows
}
