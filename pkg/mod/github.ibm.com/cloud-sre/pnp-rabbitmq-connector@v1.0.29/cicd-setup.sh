#!/bin/bash
echo
echo "---> Building rabbitmq image"
docker build -f Dockerfile.rabbitmq -t pnp-rabbitmq .
echo
echo "---> Starting rabbitmq image"
docker run -d -p 5672:5672 --hostname localhost pnp-rabbitmq
