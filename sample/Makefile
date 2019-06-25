run:
	@docker run -it \
	-v ${PWD}/envoy:/etc/envoy envoyproxy/envoy-alpine:v1.10.0 \
	-c /etc/envoy/envoy.yaml \
	-l debug \
	--service-node test-node-abcdef \
	--service-cluster test-node-cluster