package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Spec struct {
	ConfigServerAddress string `mapstructure:"CONFIG_SERVER_ADDRESS"`
}

var Config Spec

func panicEmpty(c string, configName string) {
	if c == "" {
		panic(
			fmt.Errorf(
				fmt.Sprintf("Error reading config, %s is not set.", configName),
			),
		)
	}
}

func validate() {
	panicEmpty(Config.ConfigServerAddress, "CONFIG_SERVER_ADDRESS")
}

func Read() {
	viper.BindEnv("CONFIG_SERVER_ADDRESS")

	if err := viper.Unmarshal(&Config); err != nil {
		panic(err)
	}

	validate()
}
