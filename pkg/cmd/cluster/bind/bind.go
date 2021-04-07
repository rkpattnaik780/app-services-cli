package bind

import (
	"context"
	"errors"

	"github.com/bf2fc6cc711aee1a0c2a/cli/internal/config"
	"github.com/bf2fc6cc711aee1a0c2a/cli/internal/localizer"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/cmd/factory"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/iostreams"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/kafka"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/logging"
	"github.com/spf13/cobra"
)

type Options struct {
	Config     config.IConfig
	Connection func(connectionCfg *connection.Config) (connection.Connection, error)
	Logger     func() (logging.Logger, error)
	IO         *iostreams.IOStreams

	kubeconfigLocation string
	namespace          string

	forceCreationWithoutAsk bool
	ignoreContext           bool
	selectedKafka           string
}

func NewBindCommand(f *factory.Factory) *cobra.Command {
	opts := &Options{
		Config:     f.Config,
		Connection: f.Connection,
		Logger:     f.Logger,
		IO:         f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "Connect your service with your application",
		Long:    "Command allows you to connect services created by connect command to your application",
		Example: "rhoas cluster bind",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.ignoreContext == true && !opts.IO.CanPrompt() {
				return errors.New(localizer.MustLocalize(&localizer.Config{
					MessageID: "flag.error.requiredWhenNonInteractive",
					TemplateData: map[string]interface{}{
						"Flag": "ignore-context",
					},
				}))
			}
			return runBind(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.kubeconfigLocation, "kubeconfig", "", "", localizer.MustLocalizeFromID("cluster.common.flag.kubeconfig.description"))
	cmd.Flags().StringVarP(&opts.namespace, "namespace", "n", "", localizer.MustLocalizeFromID("cluster.common.flag.namespace.description"))
	cmd.Flags().BoolVarP(&opts.forceCreationWithoutAsk, "yes", "y", false, localizer.MustLocalizeFromID("cluster.common.flag.yes.description"))
	cmd.Flags().BoolVarP(&opts.ignoreContext, "ignore-context", "", false, localizer.MustLocalizeFromID("cluster.common.flag.ignoreContext.description"))

	return cmd
}

func runBind(opts *Options) error {
	connection, err := opts.Connection(connection.DefaultConfigSkipMasAuth)
	if err != nil {
		return err
	}

	logger, err := opts.Logger()
	if err != nil {
		return err
	}

	cfg, err := opts.Config.Load()
	if err != nil {
		return err
	}

	// In future config will include Id's of other services
	if cfg.Services.Kafka == nil || opts.ignoreContext {
		// nolint
		selectedKafka, err := kafka.InteractiveSelect(connection, logger)
		if err != nil {

			return err
		}
		opts.selectedKafka = *selectedKafka.Id
	} else {
		opts.selectedKafka = cfg.Services.Kafka.ClusterID
	}

	api := connection.API()
	kafkaInstance, _, error := api.Kafka().GetKafkaById(context.Background(), opts.selectedKafka).Execute()

	if error.Error() != "" {
		return error
	}

	// Demo values
	if opts.namespace == "" {
		opts.namespace = "wtrocki-rhoas-demo"
	}

	appName := "nodejs-sample"

	err = RunSBOBind(*kafkaInstance.Name, opts.namespace, appName)

	return err
}
