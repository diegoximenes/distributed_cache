# Manual e2e test guideline

- Run services:

```bash
docker-compose --file ./docker-compose.yml up  
```

- Wait until services are ready, proxies are the ones that usually take longer to be ready.

- Define the following function.

This is useful to do requests from the host using container names, but if you prefer you can also get
the port mappings defined in `./docker-compose.yml` and use localhost with those ports instead.

```bash
dockerIP() {
    local container_name="$1"
    ip="$(docker inspect --format "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{break}}{{end}}" "$container_name")"
    echo "$ip"
}
```

- Setup nodesmedata raft cluster by adding nodesmetadata1 and nodesmetadata2 to raft cluster defined by nodesmetadata0:

```bash
curl -X PUT -i $(dockerIP nodesmetadata0):8080/raft/node -d '{"id": "raft1", "address": "nodesmetadata1:8008"}'
curl -X PUT -i $(dockerIP nodesmetadata0):8080/raft/node -d '{"id": "raft2", "address": "nodesmetadata2:8008"}'
```

- As a sanity check, check the raft cluster metadata state:

```base
curl -X GET -i -L $(dockerIP nodesmetadata0):8080/raft/metadata
```

- As a sanity check, open a new terminal to check for incremental changes on the available nodes in the distributed cache through SSE.

```bash
curl -X GET -i $(dockerIP nodesmetadata0):8080/nodes/sse
```

- Add cache nodes to the distributed cache.
Check that it works by communicating with different nodesmetadata nodes, non leader nodes will respond a redirect:

```bash
curl -X PUT -i -L $(dockerIP nodesmetadata0):8080/nodes -d '{"id": "node0", "address": "node0:8080"}'
curl -X PUT -i -L $(dockerIP nodesmetadata1):8080/nodes -d '{"id": "node1", "address": "node1:8080"}'
curl -X PUT -i -L $(dockerIP nodesmetadata2):8080/nodes -d '{"id": "node2", "address": "node2:8080"}'
```

- Add some data to the distributed cache using different proxies:

```bash
for i in $(seq 0 4); do
    echo "PUT $i"
    curl -X PUT -i $(dockerIP proxy0):8080/cache -d '{"key": "key'"$i"'", "value": "value'"$i"'"}'
done 
for i in $(seq 5 9); do
    echo "PUT $i"
    curl -X PUT -i $(dockerIP proxy1):8080/cache -d '{"key": "key'"$i"'", "value": "value'"$i"'"}'
done 
```

- Get data from the distributed cache.
Check that a key always hit the same cache node, independently of the proxy.

```bash
curl -X GET -i $(dockerIP proxy0):8080/cache/key0 
curl -X GET -i $(dockerIP proxy0):8080/cache/key0 
curl -X GET -i $(dockerIP proxy1):8080/cache/key0 
curl -X GET -i $(dockerIP proxy1):8080/cache/key0 

curl -X GET -i $(dockerIP proxy0):8080/cache/key5 
curl -X GET -i $(dockerIP proxy0):8080/cache/key5 
curl -X GET -i $(dockerIP proxy1):8080/cache/key5 
curl -X GET -i $(dockerIP proxy1):8080/cache/key5
```

- Remove a node from the distributed cache.

```bash
curl -X DELETE -i -L $(dockerIP nodesmetadata0):8080/nodes/node0
```

- Check that all keys that mapped to the removed node are not found anymore,
and hit a different node than the one removed.
Also, keys of nodes that were not removed are still hitting the same nodes.

```bash
for i in $(seq 0 9); do
    echo "GET $i"
    curl -X GET -i $(dockerIP proxy0):8080/cache/key$i 
done
```

- Stop nodesmetadata leader, now supposing that is nodesmetadata0.
Check that a new leader gets elected.

```bash
docker-compose stop nodesmetadata0
```

- Check that nodesmetadata is still operational, since 2 out of 3 nodes are still working.

```bash
curl -X GET -i -L $(dockerIP nodesmetadata1):8080/nodes
curl -X GET -i -L $(dockerIP nodesmetadata2):8080/nodes
```

- Check that the distributed cache is still operational.

```bash
curl -X PUT -i $(dockerIP proxy0):8080/cache -d '{"key": "key10", "value": "value10"}'
curl -X PUT -i $(dockerIP proxy1):8080/cache -d '{"key": "key11", "value": "value11"}'
curl -X PUT -i $(dockerIP proxy1):8080/cache -d '{"key": "key12", "value": "value12"}'

curl -X GET -i $(dockerIP proxy0):8080/cache/key10 
curl -X GET -i $(dockerIP proxy0):8080/cache/key11
curl -X GET -i $(dockerIP proxy1):8080/cache/key12 
```

- Check that raft metadata still has nodesmetadata0 associated with it.

```bash
curl -X GET -i -L $(dockerIP nodesmetadata1):8080/raft/metadata
```

- Stop another nodesmetadata node.

```bash
docker-compose stop nodesmetadata1
```

- Check that nodesmetadata is not properly working anymore.

```bash
curl -X GET -i -L $(dockerIP nodesmetadata2):8080/nodes
```

- Restart nodesmetadata1 and check a leader wining an election.

```bash
docker-compose start nodesmetadata1
```

- Check that nodesmetadata is operational again.

```bash
curl -X GET -i -L $(dockerIP nodesmetadata2):8080/nodes
```

- Restart nodesmetadata0 and check that failed contact messages stops to appear:

```bash
docker-compose start nodesmetadata0
```
