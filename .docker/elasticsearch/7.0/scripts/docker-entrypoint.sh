#! /usr/bin/bash

# Start Cerebro
/usr/share/elasticsearch/cerebro/bin/cerebro >> /usr/share/elasticsearch/cerebro/logs/cerebro.log 2>&1 &

# Start ES
/usr/share/elasticsearch/bin/elasticsearch