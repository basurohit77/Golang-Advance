#!/bin/bash
echo "---> Cleaning up rabbitmq container"
for container_id in $(docker ps -a --filter="ancestor=pnp-rabbitmq" -q);do docker stop $container_id && docker rm $container_id;done
