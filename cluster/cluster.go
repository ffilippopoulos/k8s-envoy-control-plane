package cluster

type EndpointsSet map[string]Endpoint

type Endpoint struct {
	podName string
	podIP   string
}

type Cluster struct {
	name      string
	endpoints EndpointsSet
}

func (c *Cluster) GetIPs() []string {
	var IPs []string

	for _, e := range c.endpoints {
		IPs = append(IPs, e.podIP)
	}

	return IPs
}

type ClusterStore struct {
	store map[string]*Cluster
}

var exists = struct{}{}

func (cs *ClusterStore) Init() {
	cs.store = make(map[string]*Cluster)
}

// CreateOrUpdate creates or updates a cluster with the given endpoint
func (cs *ClusterStore) CreateOrUpdate(clusterName, podName, podIP string) {

	// If cluster does not exist add it
	if c, ok := cs.store[clusterName]; !ok {
		new := &Cluster{
			name: clusterName,
		}
		new.endpoints = make(EndpointsSet)
		new.endpoints[podName] = Endpoint{
			podName: podName,
			podIP:   podIP,
		}
		cs.store[clusterName] = new
	} else {

		// if exist update the endpoints list
		c.endpoints[podName] = Endpoint{
			podName: podName,
			podIP:   podIP,
		}

	}
}

// Delete removes an endpoint from a given cluster. If the cluster has no
// remaining endpoints it deletes the cluster from the store
func (cs *ClusterStore) Delete(name, podName string) {

	if c, ok := cs.store[name]; ok {
		// if the endpoint exists then delete it
		if _, ok := c.endpoints[podName]; ok {
			delete(c.endpoints, podName)
		}

		// if the endpoints list is empty delete the cluster
		if len(c.endpoints) == 0 {
			delete(cs.store, name)
		}
	}
}
