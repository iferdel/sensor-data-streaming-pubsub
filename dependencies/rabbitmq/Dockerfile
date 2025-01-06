FROM rabbitmq:3.13-management
RUN rabbitmq-plugins enable rabbitmq_mqtt

COPY rabbitmq.conf /etc/rabbitmq/
COPY definitions.json /etc/rabbitmq/

RUN cat /etc/rabbitmq/rabbitmq.conf
