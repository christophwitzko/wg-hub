package config

import (
	"errors"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustGet[T any](val T, err error) T {
	Must(err)
	return val
}

func initViper(cmd *cobra.Command) error {
	configFile := MustGet(cmd.Flags().GetString("config"))
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("wireguard-hub.yaml")
	}
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		var viperConfigNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &viperConfigNotFound) {
			return err
		}
	}
	return nil
}

func OnInitialize(log *logrus.Logger, cmd *cobra.Command) {
	if err := initViper(cmd); err != nil {
		log.Errorf("failed to load config: %v", err)
		os.Exit(1)
	}

	logLevel := viper.GetString("logLevel")
	parsedLogLevel, ok := parseLogLevel(logLevel)
	if !ok {
		log.Warnf("failed to parse log level: %s", logLevel)
	}
	log.SetLevel(parsedLogLevel)

	usedConfigFile := viper.ConfigFileUsed()
	if usedConfigFile != "" {
		log.Infof("using config: %s", usedConfigFile)
	}
}

func parseLogLevel(logLevel string) (logrus.Level, bool) {
	switch strings.ToLower(logLevel) {
	case "d", "debug":
		return logrus.DebugLevel, true
	case "i", "info":
		return logrus.InfoLevel, true
	case "w", "warn":
		return logrus.WarnLevel, true
	case "e", "error":
		return logrus.ErrorLevel, true
	case "f", "fatal":
		return logrus.FatalLevel, true
	}
	return logrus.DebugLevel, false
}

func SetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("private-key", "", "base64 encoded private key")
	cmd.PersistentFlags().Uint16("port", 9999, "port to listen on")
	cmd.PersistentFlags().String("bind-address", "", "address to bind to")
	cmd.PersistentFlags().StringArrayP("peer", "p", nil, "base64 encoded public key and allowed ips of peer (e.g. -p \"publicKey,allowedIPs\")")
	cmd.PersistentFlags().String("config", "", "config file (default is .wireguard-hub.yaml)")
	cmd.PersistentFlags().String("log-level", "debug", "log level (debug, info, warn, error, fatal)")
	cmd.PersistentFlags().String("debug-address", "", "debug address to bind on")
	cmd.PersistentFlags().SortFlags = true

	Must(viper.BindPFlag("privateKey", cmd.PersistentFlags().Lookup("private-key")))
	viper.MustBindEnv("privateKey", "PRIVATE_KEY")
	Must(viper.BindPFlag("port", cmd.PersistentFlags().Lookup("port")))
	viper.MustBindEnv("port", "PORT")
	Must(viper.BindPFlag("bindAddress", cmd.PersistentFlags().Lookup("bind-address")))
	viper.MustBindEnv("bindAddress", "BIND_ADDRESS")
	Must(viper.BindPFlag("logLevel", cmd.PersistentFlags().Lookup("log-level")))
	viper.MustBindEnv("logLevel", "LOG_LEVEL")
	Must(viper.BindPFlag("debugAddress", cmd.PersistentFlags().Lookup("debug-address")))
	viper.MustBindEnv("debugAddress", "DEBUG_ADDRESS")
}
