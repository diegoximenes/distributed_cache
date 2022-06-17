package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Spec struct {
	NodesMetadataAddress string `mapstructure:"NODES_METADATA_ADDRESS"`
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
	panicEmpty(Config.NodesMetadataAddress, "NODES_METADATA_ADDRESS")
}

func Read() {
	viper.BindEnv("NODES_METADATA_ADDRESS")

	if err := viper.Unmarshal(&Config); err != nil {
		panic(err)
	}

	validate()
}
