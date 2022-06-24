# nodesmetadata

It is a raft cluster that stores information about which nodes compose the distributed cache.
It achieves linearizable writes and reads with non-Byzantine fault tolerance.

Implementation inspired by [rqlite](https://github.com/rqlite/rqlite).

## Application API

nodesmetadata provides an HTTP API that enables to:

- Add/Remove cache nodes to/from the distributed cache.
- Query cache nodes' information, including IDs and addresses.
- Subscribe, through SSE, to events regarding changes in the cache nodes that compose the distributed cache, including addition and removal of nodes.

It also provides an HTTP API that enables to:

- Add/Remove nodesmetadata's instances to/from the raft cluster.
- Query raft cluster's information, including instances' IDs and addresses.
- Subscribe, through SSE, to events regarding some changes in the raft cluster, including leadership changes and addition/removal of instances.

This API is binded to the port defined by the flag `--application_bind_address`.
nodesmetadata also have a flag `--application_advertised_address` to handle "NAT scenarios".

## Raft API

Raft instances communicate with each other through the port defined by the flag `--raft_bind_address`.
There is also a flag `--raft_advertised_address` to handle "NAT scenarios".

To achieve linearizable reads and writes, if a nodesmetadata's client do a request to a follower nodesmetadata's node application API, then the follower response is a redirect to the leader node application API.
Therefore, a follower node must be able to get the leader's application API address.

[hashicorp/raft](https://github.com/hashicorp/raft) implementation, which is used by this module, only provides the address and ID of the leader([LeaderWithID](https://pkg.go.dev/github.com/hashicorp/raft#Raft.LeaderWithID)). 
To solve this issue, nodesmetadata demultiplexes TCP connections binded to `--raft_bind_address` based on the first byte of the connection payload.
In one case the TCP connection will be handled by the RPC API defined by hashicorp/raft, and in the other case it will be handled by an HTTP API that will respond information about the raft instance that is receiving the request, including its application address.
Then a follower is able to get the application address of the leader on the fly, by sending an HTTP request to the leader.

At some point I thought about storing the relationship between raft node ID and application address into the raft's log.
However, this could lead to inconsistencies, such as a raft instance being able to join a cluster, but not being able to apply the necessary changes to raft's log right away, creating the need of having a procedure to reconcile this state later, which can be complex.
Another workaround would be to encode this application address information into the node's ID. 
One issue with this approach, at least with the current hashicorp/raft version, is that the in this case the advertised port associated to a machine would not be allowed to change over time, since nodes' ID are immutable.
