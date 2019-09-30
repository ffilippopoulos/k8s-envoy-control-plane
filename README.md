# k8s-envoy-control-plane

This is an example implementation of an opionionated and simple Envoy control plane. It uses custom kubernetes resources to wrap envoy configurations for clusters and listeners creation.
It's goal is to support secure communication via envoy sidecars in cross cluster setups where there is a flat network between clusters.

In brief, this is how it works:
* Pods are registered to envoy clusters by the annotations (default `cluster-name.envoy.uw.io: <cluster-name>`). This allows the control
  plane to discover and route to the pods.
* Two custom resource definitions (`IngressListener`, `EgressListener`) in Kubernetes are used to define ingress and egress listeners.
* All listeners can configure their tls context using certificates fetched from kubernetes secrets.

## Deploy

Templates to deploy using `kustomize` are provided [here](./deploy).
Also there are examples to deploy the the [control-plane](./deploy/example/control-plane/) as well as examples of usage for [tcp](./deploy/example/tcp/) and [http](deploy/example/http/) configurations.
It is handy for every namespace to include a [configMap](./deploy/example/tcp/envoy-cp-config.yaml) that points envoy instances to the control plane, as this is shared by all the instances.

Along with fetching the upstream manifests, secrets that include kubeconfigs for the controlled kubernetes clusters need to be created.
An example script to generate one:
```
# Input vars
# your server name goes here
server=<server-name>:<port>
# the name of the secret containing the service account token goes here
name=<secret-name>
# namespace and context of the secret
namespace=<namespace>
context=<context>
suffix=<suffix>

ca=$(kubectl --context=${context} --namespace=${namespace} get secret/$name -o jsonpath='{.data.ca\.crt}')
token=$(kubectl --context=${context} --namespace=${namespace} get secret/$name -o jsonpath='{.data.token}' | base64 --decode)
namespace=$(kubectl --context=${context} --namespace=${namespace} get secret/$name -o jsonpath='{.data.namespace}' | base64 --decode)

echo "
current-context: kube-${suffix}
apiVersion: v1
clusters:
- cluster:
    server: ${server}
    certificate-authority-data: ${ca}
  name: kube-${suffix}
contexts:
- context:
    cluster: kube-${suffix}
    user: service-account
  name: kube-${custom}
kind: Config
users:
- name: service-account
  user:
    token: ${token}" > sa-${suffix}.kubeconfig

kubeconfig=$(cat sa-${suffix}.kubeconfig | base64 -w 0)

echo "
apiVersion: v1
kind: Secret
metadata:
  name: secret-${suffix}
  namespace: kube-system
data:
  ca.crt: ${ca}
  kubeconfig: ${kubeconfig}
" > secret-${suffix}.yaml
```

## Static Configuration

```
  -cluster-name-annotation string
        Annotation that will mark a pod as part of a cluster (default "cluster-name.envoy.uw.io")
  -log-level string
        log level trace|debug|info|warning|error|fatal|panic (default "warning")
  -sources string
        (Required) Path of the config file that keeps static sources configuration
```

### Sources configuration - control plane

Each source is a kubernetes cluster that the control plane is going to watch events from. An example config will be:

```
    [
      {
        "name": "aws",
        "kubeconfig": "/etc/aws/aws.kubeconfig",
        "listenersource": true
      },
      {
        "name": "gcp",
        "kubeconfig": "/etc/gcp/gcp.kubeconfig",
        "listenersource": false
      }
    ]

```

You need to supply a list of source and the location of the respective kubeconfig files.
`listenersource` defines the primary kubernetes cluster, meaning the cluster which the control plane will watch for listeners. It should be set to true only for 1 cluster, the same one that the control plane is deployed. That way envoy sidecars in each kubernetes cluster will talk to the local control plane to fetch config, listeners will only be defined in the same cluster where the envoy sidecar is running and clusters will be generated from all the available sources.
Losing networking between clusters means that remote cluster members will disappear but everything will keep working for the local cluster.

## Dynamic Configuration - envoy

Configuration for listeners is generated via 2 kubernetes custom resources.

### IngressListener

```
apiVersion: ingresslistener.envoy.uw.systems/v1alpha1
kind: IngressListener
metadata:
  name: <name>
spec:
  nodename: <envoy-node-name>
  listenport: 8080 # Port to listen for ingress traffic
  targetport: 80 # Port to forward on localhost
  rbac:
    cluster: <cluster> # Only accept incoming traffic from cluster
    sans: # List of SANs to accept traffic from in case of tls
    - cluster.aws.io/bob
    - cluster.aws.io/alice
  tls:
    secret: <secret-name> # kubernetes.io/tls secret that contains `tls.crt` and `tls.key` for the tls context
    validation: <secret-name> # secret that contains `ca.crt` to be used for cert validation
```
### EgressListener

```
apiVersion: egresslistener.envoy.uw.systems/v1alpha1
kind: EgressListener
metadata:
  name: <listener-name>
spec:
  nodename: <envoy-node-name>
  listenport: 8080 # Port to listen on localhost for egress traffic.
  target:
    cluster: <cluster> # The target cluster to forward traffic to
    port: 8080 # Target cluster port to send traffic
  lbpolicy: "tcp" # load balancing policy (tcp|http), default is tcp
  tls:
    secret: <secret-name> # kubernetes.io/tls secret that contains `tls.crt` and `tls.key` for the tls context
```

## Next steps
* Healthchecks
  * Define healthchecks for clusters
* Improvements
  * Code review, look at the overall design, full rewrite
  * Every listener event triggers a complete update to the snapshot cache for every node; we could examine the contents of the event
    and decide which nodes actually need updating
  * A deleted pod is only removed from a cluster when it is completely deleted, meaning pods in a terminating state remain in the
    load balancing pool
