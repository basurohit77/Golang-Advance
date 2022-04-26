#! /bin/bash

# to run this tool, follow instructions at https://github.com/IBM/openapi-validator

lint-openapi -d $1 > /tmp/linterrors.txt

echo top of /tmp/linterrors.txt
head -n 30 /tmp/linterrors.txt
