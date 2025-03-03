package cmd

import (
	"content-prep/pkg/config"
	"content-prep/pkg/logger"
	"content-prep/pkg/packager"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(newCmd)

	newCmd.Flags().StringP(config.KeySourceFolder, "p", "", "Path to the source folder")
	_ = newCmd.MarkFlagRequired(config.KeySourceFolder)
	_ = newCmd.MarkFlagDirname(config.KeySourceFolder)
	newCmd.Flags().StringP(config.KeySetupFile, "s", "", "Path to the setup file (must be inside the source folder)")
	_ = newCmd.MarkFlagRequired(config.KeySetupFile)
	_ = newCmd.MarkFlagFilename(config.KeySetupFile)
	newCmd.Flags().StringP(config.KeyOutputFolder, "o", "", "Path to the output folder")
	_ = newCmd.MarkFlagRequired(config.KeyOutputFolder)
	_ = newCmd.MarkFlagDirname(config.KeyOutputFolder)

}

var newCmd = &cobra.Command{
	Use:     "new",
	Short:   "creates a new intunewin package from a setup file",
	Example: "content-prep new --path /path/to/folder --setupFile setup.exe --output /path/to/output",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		log := logger.FromContext(ctx).With("component", "cli", "action", "new")

		sourceFolder := viper.GetString(config.KeySourceFolder)
		setupFile := viper.GetString(config.KeySetupFile)
		outputFolder := viper.GetString(config.KeyOutputFolder)

		if !path.IsAbs(sourceFolder) {
			wd, err := os.Getwd()
			if err != nil {
				return errors.Wrapf(err, "failed to get working directory")
			}
			viper.Set(config.KeySourceFolder, path.Join(wd, sourceFolder))
		}

		if !path.IsAbs(setupFile) {
			wd, err := os.Getwd()
			if err != nil {
				return errors.Wrapf(err, "failed to get working directory")
			}
			viper.Set(config.KeySetupFile, path.Join(wd, setupFile))
		}

		if !strings.HasPrefix(setupFile, sourceFolder) {
			return errors.New("setup file must be inside the source folder")
		}

		if !path.IsAbs(outputFolder) {
			wd, err := os.Getwd()
			if err != nil {
				return errors.Wrapf(err, "failed to get working directory")
			}
			viper.Set(config.KeyOutputFolder, path.Join(wd, outputFolder))
		}

		if strings.HasPrefix(outputFolder, sourceFolder) {
			return errors.New("output folder must not be inside the source folder")
		}

		if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create output folder")
		}

		setupFileName := path.Base(setupFile)
		setupFileExt := path.Ext(setupFileName)
		packageName := strings.TrimSuffix(setupFileName, setupFileExt) + packager.PackageFileExtension

		outputFile, err := os.Create(path.Join(outputFolder, packageName))
		if err != nil {
			return errors.Wrapf(err, "failed to create output file")
		}

		source := os.DirFS(sourceFolder)

		log.Info("trying to create intunewin package", "setupFile", setupFile, "outputFolder", outputFile)

		return errors.Wrap(
			packager.Default.CreatePackage(ctx, source, setupFile, outputFile),
			"failed to create intunewin package",
		)
	},
}
