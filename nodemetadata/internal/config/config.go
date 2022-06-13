package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Spec struct {
	RaftId               string `mapstructure:"raft_id" validate:"required"`
	RaftAddress          string `mapstructure:"raft_address" validate:"required"`
	RaftDir              string `mapstructure:"raft_dir" validate:"required"`
	BootstrapRaftCluster *bool  `mapstructure:"bootstrap_raft_cluster" validate:"required"`
	HTTPAddress          string `mapstructure:"http_address" validate:"required"`
}

var Config Spec

func Read() {
	pflag.String("raft_id", "", "raft node id")
	pflag.String("raft_address", "", "raft bind address")
	pflag.String("raft_dir", "", "raft dir path")
	pflag.Bool("bootstrap_raft_cluster", false, "boolean indicating if should boostrap raft cluster")
	pflag.String("http_address", "", "http bind address")
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