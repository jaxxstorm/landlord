package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliConfig struct {
	APIURL     string
	ConfigFile string
}

var cfg cliConfig
var v *viper.Viper

func bindCLIFlags(cmd *cobra.Command) error {
	v = viper.New()

	if err := v.BindPFlag("config", cmd.PersistentFlags().Lookup("config")); err != nil {
		return err
	}
	if err := v.BindPFlag("api.url", cmd.PersistentFlags().Lookup("api-url")); err != nil {
		return err
	}
	return nil
}

func loadCLIConfig(cmd *cobra.Command) error {
	if v == nil {
		return fmt.Errorf("viper not initialized")
	}

	v.SetEnvPrefix("LANDLORD_CLI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("api.url", "http://localhost:8081")

	configFile := v.GetString("config")
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("landlord-cli")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/landlord")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config: %w", err)
		}
	}

	cfg = cliConfig{
		APIURL:     v.GetString("api.url"),
		ConfigFile: v.GetString("config"),
	}

	if cfg.APIURL == "" {
		return fmt.Errorf("api url is required")
	}

	return nil
}
