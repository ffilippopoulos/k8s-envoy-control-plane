# k8s-envoy-control-plane

This is an example implementation of a simple Envoy control plane.

In brief, this is how it works:
* Pods are registered to clusters by the annotation `cluster-name.envoy.uw.io: <cluster-name>`. This allows the control
  plane to discover and route to the pods.
* Two custom resource definitions (`IngressListener`, `EgressListener`) in Kubernetes are used to define ingress and egress listeners.

## Quick Demo
Create two separate kubeconfig files, one for `exp-1-aws` and another for `exp-1-gcp`.

Copy `config.example.json` to `config.json` and update the paths to match your kubeconfig files.
```
$ cp config.example.json config.json
$ cat config.json
[
  {
    "name": "aws",
    "kubeconfig": "/Users/robertbest/tempconfig/aws-kubeconf"
  },
  {
    "name": "gcp",
    "kubeconfig": "/Users/robertbest/tempconfig/gcp-kubeconf"
  }
]
```

Create the CRDs
```
$ cd manifests/crds/
$ kubectl --context=exp-1-aws -n labs apply -f .
```

Start the control plane
```
$ make run
2019/06/27 08:57:23 [INFO] starting pod watcher
2019/06/27 08:57:23 [INFO] starting pod watcher
2019/06/27 08:57:23 [INFO] starting ingressListener watcher
2019/06/27 08:57:23 [INFO] starting egressListener watcher
2019/06/27 08:57:23 [INFO] starting egressListener watcher
2019/06/27 08:57:23 [INFO] starting ingressListener watcher
```

Start a local instance of envoy that uses the control plane
```
$ make local-envoy
$ make local-envoy-mac # on macOS
```

You can inspect the current Envoy config by visting `/config_dump`
```
$ curl http://localhost:9901/config_dump
```

Create a pod in the cluster 'test-cluster', an `IngressListener` for the app 'test-app' and an `EgressListener` that allows
'test-app' to query `test-cluster:8080` by making requests to envoy at `127.0.0.1:9090`.
```
$ cd manifests/examples
$ kubectl --context=exp-1-aws -n labs apply -f .
```

You should see events registered in the logs of the control plane
```
2019/06/27 09:16:45 [DEBUG] received ADDED event for cluster test-cluster ip: 10.2.7.139
```
```
2019/06/27 09:20:45 [DEBUG] received ADDED event for ingress listener example-listener: 0.0.0.0:8080 -> 127.0.0.1:8081
2019/06/27 09:20:45 [DEBUG] Updating snapshot for node test-app
2019/06/27 09:20:45 [DEBUG] Response: responding (2) with type: type.googleapis.com/envoy.api.v2.Cluster, version: 2019-06-27 01:20:45.705178 +0100 BST m=+240.565251561, resources: 1
2019/06/27 09:20:45 [DEBUG] Response: responding (1) with type: type.googleapis.com/envoy.api.v2.Listener, version: 2019-06-27 01:20:45.705185 +0100 BST m=+240.565257900, resources: 1
```

Check `http://localhost:9901/config_dump` to see the listeners and clusters that have been added to Envoy by the control plane.

Add another pod, to gcp this time, to see the cluster update:
```
$ kubectl --context=exp-1-gcp -n labs apply -f clusterpod.yaml
```

## How it works
### IngressListener
```yaml
apiVersion: ingresslistener.envoy.uw.systems/v1alpha1
kind: IngressListener
metadata:
  name: example-listener
spec:
  nodename: test-app
  listenport: 8080
  targetport: 8081
  rbacallowcluster: test-cluster
```
The `IngressListener` CRD defines a listener on `0.0.0.0:listenport` which routes requests to a local cluster defined as `127.0.0.1:targetport`. Access is
restricted by an RBAC filter to source IPs belonging to pods in the cluster defined by `rbacallowcluster`.

An `IngressListener` only applies to envoys which register with a node name matching the `nodename` field.

### EgressListener
```yaml
apiVersion: egresslistener.envoy.uw.systems/v1alpha1
kind: EgressListener
metadata:
  name: example-egress-listener
spec:
  nodename: test-app
  listenport: 9090
  targetport: 8080
  targetcluster: test-cluster
```
An `EgressListener`creates a local listener on `127.0.0.1` that routes traffic to the cluster defined by `targetcluster` on
`targetport`.

An `EgressListener` only applies to envoys which register with a node name matching the `nodename` field.

## Next steps
* Create a full example running cross-cluster in kube
* mTLS
  * Define a TLS context for ingress listeners and egress clusters
  * Source certificates dynamically from a 'Secret Discovery Service' (cert-manager?)
* Healthchecks
  * Define healthchecks for clusters
* Improvements
  * Code review, look at the overall design, full rewrite
  * Every listener event triggers a complete update to the snapshot cache for every node; we could examine the contents of the event
    and decide which nodes actually need updating
  * A deleted pod is only removed from a cluster when it is completely deleted, meaning pods in a terminating state remain in the
    load balancing pool
