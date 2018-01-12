# proto
This folder contains gRPC proto files and their generated language defintions.

You can use protoc to compile these on your own:
`protoc -I ./ ./plugin.proto --go_out=plugins=grpc:./`