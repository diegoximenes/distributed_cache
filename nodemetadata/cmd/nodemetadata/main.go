package main

import (
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/httprouter"
)

func main() {
	config.Read()
	httprouter.Set()
}
