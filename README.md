# isglb

islb? is-graph-lb!

## What is this?

A library for dion:

* status sync (e.g. track route status)
* report upload (e.g. webrtc stats)
* route control

## How to use

1. Implement `Algorithm` with your route algorithm
2. Implement `Status`, `Report`, `Request` with some protocol (e.g. gRPC)
3. For server, implement `ServerConn` with some protocol (e.g. gRPC)
3. For client, implement `ClientStreamFactory` with some protocol (e.g. gRPC)

## How to implement a server

1. Start a server in your protocol (e.g. gRPC)
2. Create a `ISGLBService` by `NewISGLBService`
3. When established a new connection, call `SyncSFU`

## How to implement a client

1. Create a `Client` by `NewClient`
2. Write a `OnStatusRecv` with your protocol (e.g. gRPC)
3. Call `SendReport` and `SendStatus` as you like

## log

`Client.Logger` & `Service.Logger`