package cmd

import (
	"fmt"
	"os"

	"github.com/eblechschmidt/nixhome/internal/server"
	"github.com/spf13/cobra"
)

var cfgFile string
var dataDir string

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "c", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringVar(&dataDir, "dataDir", "d", "directory where the icon chache is stored")
}

var rootCmd = &cobra.Command{
	Use:   "nixhome",
	Short: "nixhome is a small and fast homelab homepage",
	Long:  `nixhome is a small web server serving a homepage for your homelab`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := os.Stat(cfgFile)
		if os.IsNotExist(err) {
			return err
		}
		s, err := server.New(cfgFile, ":8080")
		if err != nil {
			return err
		}

		return s.Serve()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
