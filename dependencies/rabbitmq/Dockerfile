FROM rabbitmq:4.1.0-management
RUN rabbitmq-plugins enable rabbitmq_mqtt
RUN rabbitmq-plugins enable rabbitmq_stream
RUN rabbitmq-plugins enable rabbitmq_stream_management
RUN rabbitmq-plugins enable rabbitmq_prometheus

COPY rabbitmq.conf /etc/rabbitmq/
COPY definitions.json /etc/rabbitmq/

RUN cat /etc/rabbitmq/rabbitmq.conf
