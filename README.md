# Bonos

Work in Progress.

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

New pass, with 10 uses. The pass will expire after bein used 10 times.

```
curl -X POST localhost:65139/twirp/co.rbel.bonos.v1.Bonos/New -H 'Content-Type: application/json' -d '{"amount": 10}'
```

### Use one

Use the pass once.

```
curl -X POST localhost:65139/twirp/co.rbel.bonos.v1.Bonos/Use -H 'Content-Type: application/json' -d '{}'
```

### Print used passes

Print the remaining times the pass can be used.

```
curl -X POST localhost:65139/twirp/co.rbel.bonos.v1.Bonos/Details -H 'Content-Type: application/json' -d '{}'
```

### Print pass stamps

Every time the pass is used, we add a timestamp.

```
curl -X POST localhost:65139/twirp/co.rbel.bonos.v1.Bonos/Stamps -H 'Content-Type: application/json' -d '{}'
```
