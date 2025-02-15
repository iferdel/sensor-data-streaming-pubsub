#!/bin/sh
# wait-for-rabbitmq-postgresql.sh

RABBITMQ_HOST=${RABBITMQ_HOST:-rabbitmq}
RABBITMQ_PORT=${RABBITMQ_PORT:-5672}
POSTGRES_HOST=${POSTGRES_HOST:-timescaledb}
POSTGRES_PORT=${POSTGRES_PORT:-5432}

until nc -z -v -w30 $RABBITMQ_HOST $RABBITMQ_PORT
do
  echo "Waiting for RabbitMQ..."
  sleep 1
done
echo "RabbitMQ is up and running!"

until nc -z -v -w30 $POSTGRES_HOST $POSTGRES_PORT
do
  echo "Waiting for PostgreSQL..."
  sleep 1
done
echo "PostgreSQL is up and running!"

exec "$@"
