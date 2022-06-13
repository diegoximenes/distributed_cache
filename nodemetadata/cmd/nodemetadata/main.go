package main

import (
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/config"
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/httprouter"
	"github.com/diegoximenes/distributed_key_value_cache/nodemetadata/internal/raft"
)

func main() {
	config.Read()

	raftNode, fsm, err := raft.Set()
	if err != nil {
		panic(err)
	}

	httprouter.Set(raftNode, fsm)
}