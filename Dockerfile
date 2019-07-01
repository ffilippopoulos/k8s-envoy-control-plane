FROM golang:1.12-alpine AS build

ENV GOPATH=/go

WORKDIR /go/src/github.com/ffilippopoulos/k8s-envoy-control-plane
COPY . /go/src/github.com/ffilippopoulos/k8s-envoy-control-plane

ENV CGO_ENABLED 0
RUN apk --no-cache add ca-certificates git && \
  go get -u github.com/golang/dep/cmd/dep && \
  /go/bin/dep ensure && \
  go test ./... && \
  (cd cmd/ && go build -o /k8s-envoy-control-plane .)

FROM alpine:3.10

COPY --from=build /k8s-envoy-control-plane /k8s-envoy-control-plane

ENTRYPOINT [ "/k8s-envoy-control-plane" ]
