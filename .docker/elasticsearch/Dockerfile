FROM docker.elastic.co/elasticsearch/elasticsearch:7.0.1

USER elasticsearch

# Copy config files
COPY ./jvm.options /usr/share/elasticsearch/config/
COPY ./log4j2.properties /usr/share/elasticsearch/config/
COPY ./elasticsearch.yml /usr/share/elasticsearch/config/

# Install Cerebro
ARG CEREBRO_VERSION
RUN cd /usr/share/elasticsearch/ \
    && wget -O cerebro-${CEREBRO_VERSION}.tgz https://github.com/lmenezes/cerebro/releases/download/v${CEREBRO_VERSION}/cerebro-${CEREBRO_VERSION}.tgz \
    && tar zxvf cerebro-${CEREBRO_VERSION}.tgz \
    && rm cerebro-${CEREBRO_VERSION}.tgz \
    && mkdir cerebro-${CEREBRO_VERSION}/logs \
    && mv cerebro-${CEREBRO_VERSION} cerebro

COPY ./docker-entrypoint.sh /usr/share/elasticsearch/start