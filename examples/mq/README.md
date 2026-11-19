# Kafka worker pool example

This standalone Go program demonstrates a producer and a competing-consumer worker pool. Each worker owns a Kafka consumer in the same consumer group, so Kafka assigns topic partitions across the workers. A message is committed only after its handler succeeds.

## Start Kafka

From this directory, start the single-node Kafka broker:

```powershell
docker compose up -d
```

Create the example topic with three partitions:

```powershell
docker compose exec broker /opt/kafka/bin/kafka-topics.sh --create --if-not-exists --topic jobs --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

## Run

In one terminal, start three workers:

```powershell
go run . worker --workers 3
```

In another terminal, publish ten jobs:

```powershell
go run . produce --count 10
```

Stop the workers with Ctrl+C, and stop Kafka when finished:

```powershell
docker compose down
```

## Configure

Both commands accept `--brokers` and `--topic`. The worker command also accepts `--group` and `--workers`; the producer accepts `--count`. They can also be set with `KAFKA_BROKERS`, `KAFKA_TOPIC`, and `KAFKA_GROUP`.

For example:

```powershell
go run . worker --brokers localhost:9092 --topic jobs --group example-workers --workers 3
```

With one broker, the three topic partitions can be handled concurrently by up to three active workers. Additional workers in the group wait without an assigned partition. Kafka consumer groups distribute each partition to only one active worker at a time.

If processing fails, the worker pool stops without committing that message. Restarting the same group allows Kafka to deliver uncommitted work again. This is at-least-once delivery: production handlers should be idempotent because a process can fail after performing the work but before committing its offset.
