package cmd

import (
	"content-prep/pkg/config"
	"content-prep/pkg/logger"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(func() {
		viper.SetEnvPrefix("CONTENT_PREP")
		viper.AutomaticEnv()

		walkBindCommands([]*cobra.Command{RootCmd})
	})

	RootCmd.PersistentFlags().Bool(config.KeyJSONLogging, false, "enable JSON logging (CONTENT_PREP_JSON)")
	RootCmd.PersistentFlags().Bool(config.KeyVerboseLogging, false, "enable verbose logging (CONTENT_PREP_VERBOSE)")
}

func walkBindCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		_ = viper.BindPFlags(cmd.Flags())
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
				_ = cmd.Flags().Set(f.Name, viper.GetString(f.Name))
			}
		})

		if cmd.HasSubCommands() {
			walkBindCommands(cmd.Commands())
		}
	}
}

var RootCmd = &cobra.Command{
	Use:              "content-prep",
	Short:            "open-source implementation of the Microsoft ContentPrep tool",
	TraverseChildren: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		l := logger.Init(viper.GetBool(config.KeyJSONLogging), viper.GetBool(config.KeyVerboseLogging))

		ctx := cmd.Context()
		ctx = logger.IntoContext(ctx, l)

		cmd.SetContext(ctx)
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		l := logger.FromContext(RootCmd.Context())

		l.Error("error executing command", "error", err)
		os.Exit(1)
	}
}
