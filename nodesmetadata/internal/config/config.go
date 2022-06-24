package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Spec struct {
	RaftID                       string `mapstructure:"raft_id" validate:"required"`
	RaftBindAddress              string `mapstructure:"raft_bind_address" validate:"required"`
	RaftAdvertisedAddress        string `mapstructure:"raft_advertised_address" validate:"required"`
	RaftDir                      string `mapstructure:"raft_dir" validate:"required"`
	BootstrapRaftCluster         bool   `mapstructure:"bootstrap_raft_cluster" validate:"required"`
	ApplicationBindAddress       string `mapstructure:"application_bind_address" validate:"required"`
	ApplicationAdvertisedAddress string `mapstructure:"application_advertised_address" validate:"required"`
}

var Config Spec

func Read() {
	pflag.String("raft_id", "", "raft node id")
	pflag.String("raft_bind_address", "", "raft bind address")
	pflag.String("raft_advertised_address", "", "raft advertised address")
	pflag.String("raft_dir", "", "raft dir path")
	pflag.Bool("bootstrap_raft_cluster", false, "bool indicating if should boostrap raft cluster")
	pflag.String("application_bind_address", "", "application bind address")
	pflag.String("application_advertised_address", "", "application advertised address")
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
