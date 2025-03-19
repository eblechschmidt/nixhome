package cmd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/eblechschmidt/nixhome/config"
	"github.com/eblechschmidt/nixhome/internal/server"
	"github.com/eblechschmidt/nixhome/web"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cfgFile string
var dataDir string
var addr string

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringVar(&dataDir, "dataDir", "", "directory where the icon chache is stored")
	rootCmd.PersistentFlags().StringVar(&addr, "addr", "", "address the web server should bind to (e.g. :8080)")
	_ = rootCmd.MarkFlagRequired("config")
	_ = rootCmd.MarkFlagRequired("dataDir")
	_ = rootCmd.MarkFlagRequired("addr")
}

var rootCmd = &cobra.Command{
	Use:   "nixhome",
	Short: "nixhome is a small and fast homelab homepage",
	Long:  `nixhome is a small web server serving a homepage for your homelab`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := ensureTemplate(); err != nil {
			return err
		}
		s, err := server.New(cfgFile, addr, dataDir)
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

func ensureConfig() error {
	_, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		log.Info().Str("cfgFile", cfgFile).Msg("Config file does not exist. Create one.")
		err := os.WriteFile(cfgFile, config.Example, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create sample config: %w", err)
		}
	}
	return nil
}

func ensureTemplate() error {
	dir := filepath.Join(dataDir, "template")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	items, err := fs.ReadDir(web.FS, ".")
	if err != nil {
		return err
	}
	for _, i := range items {
		if i.IsDir() {
			continue
		}
		fname := filepath.Join(dir, i.Name())
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			from, err := web.FS.Open(i.Name())
			if err != nil {
				return err
			}
			defer from.Close()

			to, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				return err
			}
			defer to.Close()

			_, err = io.Copy(to, from)
			if err != nil {
				return err
			}
			log.Info().Str("file", fname).Msg("File created")
		}
	}
	return nil
}
