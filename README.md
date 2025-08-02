# Pubbet

Pubbet is a lightweight and fast message queue.

Written in Go using gRPC under the hood.

## Running

```shell
docker build -t pubbet .
docker run -p 5000:5000 -d pubbet
```

## SDKs

There is a simple [library](https://github.com/misshanya/pubbet-sdk-go) for Go to interact with Pubbet.