Test using curl:

```bash
go run cmd/server/main.go
curl -X POST localhost:8080 -d '{"record": {"value":"TGV0J3MgR28gIzEK"}}'
curl -X GET localhost:8080 -d '{"offset": 0}'
```