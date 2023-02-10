# Introduction

`otel-demo` is a demo microservice structure app for opentelemetry. 

It has three components:
- client: do user request randomly to server
- server: query username, then request worker to `say hello`
- worker: response `hello {username}`

# Build & Run & Deploy
## build
```
make build
```
## image
```
make docker-build docker-push
```
## deploy to k8s
```
make deploy
```
