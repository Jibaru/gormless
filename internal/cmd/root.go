package cmd

import (
	"fmt"
	"os"

	"github.com/Jibaru/gormless/internal/generator"
	"github.com/Jibaru/gormless/internal/parser"
	"github.com/spf13/cobra"
)

var (
	input  string
	output string
	driver string
)

var rootCmd = &cobra.Command{
	Use:   "gormless",
	Short: "Generate DAOs for Go models",
	Long:  `Gormless is a tool that helps Go developers generate DAOs for models in their code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateFlags(); err != nil {
			return err
		}

		if err := checkInputExists(); err != nil {
			return err
		}

		models, err := parser.ParseModels(input)
		if err != nil {
			return err
		}

		if len(models) == 0 {
			return fmt.Errorf("there are no models")
		}

		return generator.GenerateDAOs(models, output, driver)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&input, "input", "i", "", "Path to file or folder with models (required)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Path to output folder (required)")
	rootCmd.Flags().StringVarP(&driver, "driver", "d", "", "Database driver: postgres, mysql, sqlserver, oracle (required)")

	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("output")
	rootCmd.MarkFlagRequired("driver")
}

func Execute() error {
	return rootCmd.Execute()
}

func validateFlags() error {
	if input == "" {
		return fmt.Errorf("input not provided")
	}
	if output == "" {
		return fmt.Errorf("output folder not provided")
	}
	if driver == "" {
		return fmt.Errorf("driver not provided")
	}
	if driver != "postgres" && driver != "mysql" && driver != "sqlserver" && driver != "oracle" {
		return fmt.Errorf("invalid driver: %s. Allowed drivers: postgres, mysql, sqlserver, oracle", driver)
	}
	return nil
}

func checkInputExists() error {
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input not exists")
	}
	return nil
}