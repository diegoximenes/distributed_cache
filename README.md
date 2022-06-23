# distributed_cache

A distributed key-value cache.

This was developed so I could get a bit of implementation experience on some techniques, such as:

- Consensus through raft.
- Key partitioning through rendezvous hashing and consistent hashing.
- Connection multiplexing based on payload in golang.

It is not a production ready software.

## Project structure

`/node` represents a cache instance of the distributed cache.

`/nodesmetadata` it is a raft cluster that stores information about which nodes compose the distributed cache.

`/proxy` clients of the distributed cache communicate with proxies instead of
directly connecting to nodes.
Proxies select nodes based on the key provided by clients and the deployed
key partitioning strategy.
Proxies use nodesmetadata to get node's information.

`/util` is a module with functionality shared across the other modules.

`/test` contains a guideline on how to run manual e2e tests on this project.

The README.md placed in these directories give a more detailed description of those components.

## Future work

- Today adding or removing instances is done through manual triggers.
One idea is to automate instances' health and system's load monitoring,
to then automated the process of creating and removing instances.
This should be applied for different kinds of instances, `node`'s instances,
`nodesmetadata`'instances, and `proxies`'s instances, which will likely require
different strategies.
- Use circuit-breaker, rate-limit, and other resilience communication patterns.
- Use TLS, at least in some APIs.
- Improve load uniformity in nodes when using consistent hashing.
