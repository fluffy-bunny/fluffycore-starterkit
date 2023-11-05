# fluffycore-starterkit

starter kit for a fluffycore DI based application

## Protos

Note: I had to run bash on windows so I could pass ```./api/proto/**/*.proto```  

```bash
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

```bash
go get github.com/fluffy-bunny/fluffycore   
go install github.com/fluffy-bunny/fluffycore/protoc-gen-go-fluffycore-di/cmd/protoc-gen-go-fluffycore-di@latest

protoc --go_out=. --go_opt paths=source_relative --grpc-gateway_out . --grpc-gateway_opt paths=source_relative --openapiv2_out=allow_merge=true,merge_file_name=proto:./proto --go-grpc_out . --go-grpc_opt paths=source_relative --go-fluffycore-di_out .  --go-fluffycore-di_opt paths=source_relative,grpc_gateway=true  ./proto/helloworld/helloworld.proto  
```

## Private OAuth2 server

The kit comes with a self contained oauth2 server.  

Your apis need tokens, and [here](./cmd/server/config/client.json) we can define exactly what claims a given client will mint.  

The client_credenitials flow is the only thing supported.  

[discovery](http://localhost:50053/.well-known/openid-configuration)  
[jwks](http://localhost:50053/.well-known/jwks.json)  

client_credentials example:  

```bash
curl --location 'http://localhost:50053/oauth/token' --header 'Content-Type: application/x-www-form-urlencoded' --header 'Authorization: Basic Y2xpZW50MTpzZWNyZXQ=' --data-urlencode 'grant_type=client_credentials'
```

```json
{
    "access_token": "eyJhbGciOiJFUzI1NiIsImtpZCI6IjBiMmNkMmU1NGM5MjRjZTg5ZjAxMGYyNDI4NjIzNjdkIiwidHlwIjoiSldUIn0.eyJjbGllbnRfaWQiOiJjbGllbnQxIiwiZXhwIjoxNjk5MjI3MzY3LCJpYXQiOjE2OTkyMjM3NjcsImlzcyI6Imh0dHA6Ly9sb2NhbGhvc3Q6NTAwNTMiLCJwZXJtaXNzaW9ucyI6WyJyZWFkIiwid3JpdGUiXSwic3ViIjoiY2xpZW50MSJ9.hAtAa5W81NATUZmNDVQdQLYSmA_0Wx4HvmSMOcqGMdQMS7ay99v1RmKf-kT2l8Xm6rDMG8klIiEU9M-FK-400w",
    "expires_in": 3600,
    "token_type": "Bearer"
}
```
