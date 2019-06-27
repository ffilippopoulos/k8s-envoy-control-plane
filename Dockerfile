FROM alpine:3.10

ENV GOPATH=/go

WORKDIR /go/src/github.com/ffilippopoulos/k8s-envoy-control-plane
COPY . /go/src/github.com/ffilippopoulos/k8s-envoy-control-plane

RUN apk --no-cache add ca-certificates git go musl-dev && \
  go get -u github.com/golang/dep/cmd/dep && \
  /go/bin/dep ensure && \
  go test ./... && \
  (cd cmd/ && go build -ldflags '-s -extldflags "-static"' -o /k8s-envoy-control-plane .) && \
  apk del go git musl-dev && \
  rm -rf $GOPATH
