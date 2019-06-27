VERSION=$(shell git describe --tags --match=v* --always --dirty)
REPO=quay.io/ffilippopoulos/k8s-envoy-control-plane

run:
	@bash -c "cd cmd/ && go run main.go config.go -cluster-name-annotation cluster-name.envoy.uw.io -sources ../config.json"

local-envoy:
	@docker run -it \
	--network host \
	-v ${PWD}/sample/envoy:/etc/envoy envoyproxy/envoy-alpine:v1.10.0 \
	-c /etc/envoy/envoy.yaml \
	-l debug \
	--service-node test-app \
	--service-cluster test-app-cluster

local-envoy-mac:
	@docker run -it -p 9901:9901 \
	-v ${PWD}/sample/envoy:/etc/envoy envoyproxy/envoy-alpine:v1.10.0 \
	-c /etc/envoy/envoy-mac.yaml \
	-l debug \
	--service-node test-app \
	--service-cluster test-app-cluster

docker-build:
	@docker build -t quay.io/ffilippopoulos/k8s-envoy-control-plane .

.PHONY: docker-image
docker-image:
	@docker build --rm=true -t $(REPO):$(VERSION) .

.PHONY: docker-push
docker-push: docker-image
	@docker push $(REPO):$(VERSION)
