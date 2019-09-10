VERSION=$(shell git describe --tags --match=v* --always --dirty)
REPO=quay.io/ffilippopoulos/k8s-envoy-control-plane

run:
	@bash -c "cd cmd/ && go run main.go config.go -cluster-name-annotation cluster-name.envoy.uw.io -sources ../config.example.json"

.PHONY: build-integration
build-integration:
	@docker-compose -f integration/docker-compose.yaml build

.PHONY: integration
integration:
	@docker-compose -f integration/docker-compose.yaml up --exit-code-from envoy-cp

.PHONY: docker-image
docker-image:
	@docker build --rm=true -t $(REPO):$(VERSION) .

.PHONY: docker-push
docker-push: docker-image
	@docker push $(REPO):$(VERSION)
