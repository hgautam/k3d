# Thoughts

## Repository/Package Overview

- `cmd/`: everything around the CLI of k3d = human interface, printed output (e.g. list of clusters)
- `pkg/`: everything else, can be used as a module from other Go projects
  - `cluster/`: everything around managing cluster components
  - `runtimes/`: translate k3d types (node, cluster, etc.) to container runtime specific types and manage them
  - `types/`: collection of types (structs) and constants used by k3d
  - `util/`: utilities, that could be used for everything, not directly related to the project

## k3d types <-> runtime translation

k3d _should_ work with more than one runtime, if we can implement the Runtime interface for it.
Here's how k3d types should translate to a runtime type:

- `cluster` = set of _containers_ running in the same _network_, maybe mounting the same _volume(s)_
- `node` = _container_ with _exposed ports_ and _volume mounts_

### Docker

#### Node to Container translation

`container = "github.com/docker/docker/api/types/container"`
`network = "github.com/docker/docker/api/types/network"`

- Name    -> container.Hostname = node.Name
- Role    -> container.Labels["k3d.role"] = node.Role
- Image   -> container.Image = node.Image
- Volumes -> container.HostConfig.PortBindings
- Env     -> 
- Args    -> 
- Ports   -> 
- Restart -> 
- Labels  -> container.Labels

## Node Configuration

- server node(s)
  - ENV
    - `K3S_CLUSTER_INIT`
      - if num_servers > 1 && no external datastore configured
    - `K3S_KUBECONFIG_OUTPUT`
      - k3d default -> `/output/kubeconfig.yaml`
  - CMD/ARGS
    - `--https-listen-port`
      - can/should be left default (unset = 6443), since we handle it via port mapping
      - `--tls-san=<some-ip-or-hostname>`
        - get from `--api-port` k3d flag and/or from docker machine
  - Runtime Configuration
    - nothing special
- all nodes
  - ENV
    - `K3S_TOKEN` for node authentication
  - CMD/ARGS
    - nothing special
  - Runtime Configuration
    - Volumes
      - shared image volume
        - cluster-specific (create cluster) or inherit from existing (create node)
      - tmpfs for k3s to work properly
        - `/run`
        - `/var/run`
    - Capabilities/Security Context
      - `privileged`
    - Network
      - cluster network or external/inherited
- agent nodes
  - ENV
    - `K3S_URL` to connect to server node
      - server hostname + port (6443)
      - cluster-specific or inherited
  - CMD/ARGS
    - nothing special
  - Runtime Configuration
    - nothing special

## Features

## [DONE] Node Filters

- `--port [host:]port[:containerPort][/protocol][@group_identifier[[index] | @node_identifier]`
  - Examples:
    - `--port 0.0.0.0:8080:8081/tcp@agents` -> whole group
    - `--port 80@agents[0]` -> single instance of group by list index
    - `--port 80@agents[0,2-3]` -> multiple instances of a group by index lists and ranges
    - `--port 80@k3d-test-agent-0` -> single instance by specific node identifier
    - `--port 80@k3d-test-server-0@agents[1-5]` -> multiple instances by combination of node and group identifiers

- analogous for volumes

## [WIP] Multi-Server Setup

- to make this possible, we always deploy a load-balancer (nginx) in front of the server nodes as an extra container
  - consider that in the kubeconfig file and `--tls-san`

### Variants

- [x] embedded datastore (dqlite)
  - if `--servers` > 1 deploy a load-balancer in front of them as an extra container
- [ ] external datastore

## [DONE] Keep State in Docker Labels

- when creating a cluster, usually, you also create a new docker network (and maybe other resources)
  - store a reference to those in the container labels of cluster nodes
  - when deleting the cluster, parse the labels, deduplicate the results and delete the additional resources
  - DONE for network
    - new labels `k3d.cluster.network=<ID>` and `k3d.cluster.network.external=<true|false>` (determine whether to try to delete it when you delete a cluster, since network may have been created manually)

## Bonus Ideas

### Tools

- maybe rename `k3d load` to `k3d tools` and add tool cmds there?
  - e.g. `k3d tools import-images`
  - let's you set tools container version
    - `k3d tools --image k3d-tools:v2 import-images`
- add `k3d create --image-vol NAME` flag to re-use existing image volume
  - will add `k3d.volumes.imageVolume.external: true` label to nodes
    - should not be deleted with cluster
  - possibly add `k3d create volume` and `k3d create network` to create external volumes/networks?

### Prune Command

- `k3d prune` to prune all dangling resources
  - nodes, volumes, networks

### Use Open Standards (OCI, CRI, ...)

- [https://github.com/opencontainers/runtime-spec/blob/master/specs-go/config.go](https://github.com/opencontainers/runtime-spec/blob/master/specs-go/config.go)
- move node -> container translation out of runtime

### Private registry

- create a private registry to be used by k3d clusters
  - similar to [https://github.com/rancher/k3d/pull/161](https://github.com/rancher/k3d/pull/161)
- add `k3d create registry` command to create external registry (maybe instead of flags as in PR #161?)

### Syntactical shortcuts for k3d v1 backwards compatibility

- e.g. `k3d create` -> `k3d create cluster k3s-default`

### Unsorted Ideas

- Integrate build tool (e.g. buildkit, buildah, ...)
- use `tools.go` to keep tools (like `golangci-lint` and `gox`) dependencies
  - see e.g. [https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)
  - see e.g. [https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module](https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module)

### Possible Enhancements

- [!] remove/add nodes -> needs to remove line in `/var/lib/rancher/k3s/server/cred/node-passwd` for the deleted node
