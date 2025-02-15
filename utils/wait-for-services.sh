#!/bin/sh
# wait-for-rabbitmq-postgresql.sh

# Extract RabbitMQ host and port from the connection string
RABBITMQ_HOST=$(echo $RABBIT_CONN_STRING | awk -F[@:/] '{print $6}')
RABBITMQ_PORT=$(echo $RABBIT_CONN_STRING | awk -F[@:/] '{print $7}')

# Extract PostgreSQL host and port from the connection string
POSTGRES_HOST=$(echo $POSTGRES_CONN_STRING | awk -F[@:/] '{print $6}')
POSTGRES_PORT=$(echo $POSTGRES_CONN_STRING | awk -F[@:/] '{print $7}')

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
