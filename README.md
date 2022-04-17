# OpenTelemetry Demo

## Summary

This repo contains a environment for demonstrating OpenTelemetry tracing
support in [Uptrace](https://uptrace.dev).

## The Password Generator Service

A password generator service is instrumented with [OpenTelemetry tracing](https://opentelemetry.uptrace.dev/guide/go-tracing.html). 
This is an absurd service and should not be taken as a shining example of architecture nor coding. 
It exists as a playground example to generate traces. 
The [lower service](./cmd/lower) generates random lowercase letters. 
The [upper service](./cmd/upper) service generates random uppercase letters. 
The [digit service](./cmd/digit) generates random digits, and the [special service](./cmd/special) generates random special characters. 
There is a [generator](./cmd/generator) service which makes calls to the other services to compose a random password. 
Finally, there is a [load script](./cmd/load) which continuously calls the generator service in order to simulate user load.

All the services are written in Go.

## The Observability Infrastructure

All the microservices forward their traces to an instance of the [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/).
The collector sends the traces on to an instance of the [Uptrace](https://uptrace.dev/open-source).

## Running the System

The system runs in docker, is configured via the [docker compose file](./docker-compose.yaml), and is operated with docker-compose.
Run the following command from the root of the repo to (re)start the system.

```
docker-compose up --remove-orphans --build --detach
```

or 

```
make deploy/up
```

Once running, the following links will let you explore the various components of the system:

- [password generator service](http://localhost:5050/)
- [digit service](http://localhost:5051/)
- [special service](http://localhost:5052/)
- [lower service](http://localhost:5053/)
- [upper service](http://localhost:5054/)
- [Uptrace](http://localhost:14318/)

When you are ready to shutdown the system, use the following command.

```
docker-compose down
```

or 

```
make deploy/down
```

To destroy the environment instead of stop, use the following command.

```shell
make deploy/destroy
```