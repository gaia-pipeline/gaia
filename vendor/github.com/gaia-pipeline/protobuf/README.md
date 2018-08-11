# proto
This folder contains gRPC proto files and their generated language defintions for the gaia plugin interface.

You can use protoc to compile these on your own:
`protoc -I ./ ./plugin.proto --go_out=plugins=grpc:./`
