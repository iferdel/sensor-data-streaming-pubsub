#!/bin/sh
# wait-for-rabbitmq-postgresql.sh

until nc -z -v -w30 rabbitmq 5672
do
  echo "Waiting for RabbitMQ..."
  sleep 1
done
echo "RabbitMQ is up and running!"

until nc -z -v -w30 timescaledb 5432
do
  echo "Waiting for PostgreSQL..."
  sleep 1
done
echo "PostgreSQL is up and running!"

exec "$@"
