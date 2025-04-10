Test using curl:

```bash
go run cmd/server/main.go
curl -X POST localhost:8080 -d '{"record": {"value":"TGV0J3MgR28gIzEK"}}'
curl -X GET localhost:8080 -d '{"offset": 0}'
```

Notes
1. Compiling gRPC
```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf/cmd/protoc-gen-go@latest
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
make compile
```

2. Generate certs
```bash
go install github.com/cloudflare/cfssl/cmd/cfssl@latest
go install github.com/cloudflare/cfssl/cmd/cfssljson@latest
make gencert
```