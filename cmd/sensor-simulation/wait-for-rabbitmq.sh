#!/bin/sh
# wait-for-rabbitmq.sh

until nc -z -v -w30 rabbitmq 5672
do
  echo "Waiting for RabbitMQ..."
  sleep 1
done
echo "RabbitMQ is up and running!"
exec "$@"
