ROOT_PACKAGE="github.com/ffilippopoulos/k8s-envoy-control-plane"

INGRESS_LISTENER_RESOURCE_NAME="ingresslistener"

INGRESS_LISTENER_RESOURCE_VERSION="v1alpha1"

EGRESS_LISTENER_RESOURCE_NAME="egresslistener"

EGRESS_LISTENER_RESOURCE_VERSION="v1alpha1"
#go get -u k8s.io/code-generator/...
cd $GOPATH/src/k8s.io/code-generator

./generate-groups.sh all "$ROOT_PACKAGE/pkg/client" "$ROOT_PACKAGE/pkg/apis" "$INGRESS_LISTENER_RESOURCE_NAME:$INGRESS_LISTENER_RESOURCE_VERSION $EGRESS_LISTENER_RESOURCE_NAME:$EGRESS_LISTENER_RESOURCE_VERSION"
