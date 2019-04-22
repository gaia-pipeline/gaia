# proto
This folder contains gRPC proto files and their generated language defintions for the gaia worker interface.

You can use protoc to compile these on your own:
`protoc -I ./ ./worker.proto --go_out=plugins=grpc:./`

