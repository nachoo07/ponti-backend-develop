#!/bin/bash

echo "Docker logs."
echo "modifcar archivo .yml y servico comforme sea necesario."
docker compose -f ../config/docker-compose.dev.yml logs -f greeter-client