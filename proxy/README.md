# proxy

Clients of the distributed cache communicate with proxies instead of
directly connecting to nodes.
Proxies select nodes based on the key provided by clients and the deployed
key partitioning strategy (rendezvous hashing or consistent hashing with virtual nodes).
Proxies use nodesmetadata to get node's information.

## nodesmetadata client

nodesmetadata client doesn't assume that there is a single DNS name associated with all nodesmetadata instances,
or that this DNS name is related to the nodesmetadata leader IP.
Instead, nodesmetadata client tracks the names, or IPs, of all nodesmetadata instances.
When communicating with a follower, instead of blindly following a
redirect to reach the leader instance, nodesmetadata client updates its state to
store the new nodesmetadata leader address.
Therefore, nodesmetadata client always tries to communicate with the leader,
but if eventually communicates with a follower then the client will learn the new leader address returned by this follower.
If nodesmetadata client is not able to communicate with the leader address that it has previously stored, 
then it communicates with the followers that it is aware of, until it finds the new leader.

nodesmetadata client also keeps two SSE streams connections.
One to track changes in nodesmetadata raft cluster, e.g., addition/removal of raft instances.
And the other to track changes in the cache nodes, e.g., addition/removal of cache nodes.
nodesmetadata client doesn't use those incremental information to update its state,
but uses it as a trigger to query the full state from nodesmetata service.
