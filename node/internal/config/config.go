package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Spec struct {
	CacheSize int `mapstructure:"cache_size" validate:"required"`
}

var Config Spec

func Read() {
	pflag.Int("cache_size", 0, "cache size")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: %s [options]\n", os.Args[0])
		pflag.PrintDefaults()
	}
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	if err := viper.Unmarshal(&Config); err != nil {
		panic(err)
	}

	validate := validator.New()
	if err := validate.Struct(&Config); err != nil {
		pflag.Usage()
		panic(fmt.Sprintf("Invalid config:\n%v", err))
	}
}
