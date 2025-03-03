package cmd

import (
	"content-prep/pkg/config"
	"content-prep/pkg/logger"
	"content-prep/pkg/packager"
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(decryptIntuneWinCmd)

	decryptIntuneWinCmd.Flags().StringP(config.KeyEncryptedPackageFile, "f", "", "Path to the encrypted package file")
	_ = decryptIntuneWinCmd.MarkFlagRequired(config.KeyEncryptedPackageFile)
	_ = decryptIntuneWinCmd.MarkFlagFilename(config.KeyEncryptedPackageFile)
}

var decryptIntuneWinCmd = &cobra.Command{
	Use:     "decrypt",
	Short:   "decrypts an intunewin package",
	Example: "content-prep decrypt-intunewin-package --file /path/to/package.intunewin --output /path/to/output",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		log := logger.FromContext(ctx).With("component", "cli", "action", "decrypt")

		packageFilePath := viper.GetString(config.KeyEncryptedPackageFile)

		if !path.IsAbs(packageFilePath) {
			wd, err := os.Getwd()
			if err != nil {
				return errors.Wrapf(err, "failed to get working directory")
			}
			viper.Set(config.KeyEncryptedPackageFile, path.Join(wd, viper.GetString(config.KeyEncryptedPackageFile)))
		}

		log.Info("trying to decrypt intunewin package", "file", packageFilePath)

		file, err := os.Open(packageFilePath)
		if err != nil {
			return errors.Wrapf(err, "failed to open package file")
		}

		return errors.Wrap(
			packager.Default.DecryptPackage(ctx, file, path.Dir(packageFilePath)),
			"failed to decrypt intunewin package",
		)
	},
}
