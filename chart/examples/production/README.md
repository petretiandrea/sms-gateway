# Production Example

This example deploys `sms-gateway` with:

- shared non-sensitive config from a generated ConfigMap
- `POSTGRES_DSN` and `RABBITMQ_DSN` from one existing Kubernetes Secret
- an API init container that runs database migrations and applies RabbitMQ topology
- API and outbox as separate Deployments
- one Service for the API only
- outbox-specific runtime tuning

The RabbitMQ topology:

- exchange: `sms`
- queues: `sms.send.internal`, `sms.send.dead-letter`, `sms.delivery-confirmed`
- routing keys: `sms.send.internal.requested`, `sms.send.dead-letter`, `sms.delivery.confirmed`

RabbitMQ permissions must already exist for the DSN user. The init command creates exchanges, queues, and bindings over AMQP.

## Create the existing Secret

Edit `existing-secret.yaml` with real connection strings, then apply it:

```sh
kubectl apply -f existing-secret.yaml
```

## Render the chart

```sh
helm template sms-gateway ../.. -f values.yaml
```

## Install or upgrade

```sh
helm upgrade --install sms-gateway ../.. -f values.yaml
```

## Alternative Secret key mapping

If PostgreSQL and RabbitMQ DSNs are stored in two different Secrets, use:

```sh
helm upgrade --install sms-gateway ../.. -f values-envfromkeys.yaml
```
