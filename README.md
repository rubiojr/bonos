# Bonos

Bonos is a sample Twirp API service to manage passes that expire after a number of times.

The service is a demo of how to build a simple real life application one step at a time with:

* Docker
* Protobuf/Twirp API
* HMAC Authentication
* A key/value store as the datastore

`Bono(s)` in spanish means `pass`.

## Running the service

```Go
docker run -p 65139:65139 -v $PWD/db:/data  ghcr.io/rubiojr/bonos:latest
```

## API usage

Only one pass is allowed at a time.

### Create the pass

```
curl -X POST localhost:65139/twirp/co.rbel.bonos.v1.Bonos/New -H 'Content-Type: application/json' -d '{"amount": 10}'
```

### Use one

```
curl -X POST localhost:65139/twirp/co.rbel.bonos.v1.Bonos/Use -H 'Content-Type: application/json' -d '{}'
```

### Print used passes

```
curl -X POST localhost:65139/twirp/co.rbel.bonos.v1.Bonos/Details -H 'Content-Type: application/json' -d '{}'
```
