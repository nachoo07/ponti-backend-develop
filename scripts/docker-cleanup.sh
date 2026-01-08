#!/bin/bash

# Solicitar confirmación del usuario
read -p "Para continuar con la limpieza completa de Docker, escribe 'clean up' y presiona Enter: " confirm

# Verificar la entrada del usuario
if [ "$confirm" != "clean up" ]; then
  echo "Acción cancelada. No se realizó ninguna limpieza."
  exit 0
fi

set -eux

# 1) Parar todos los contenedores en ejecución
docker ps -q | xargs -r docker stop

# 2) Eliminar todos los contenedores
docker ps -aq | xargs -r docker rm

# 3) Eliminar todas las imágenes
docker images -q | xargs -r docker rmi -f

# 4) Eliminar todos los volúmenes
docker volume ls -q | xargs -r docker volume rm -f

# 5) Eliminar todas las redes no utilizadas
docker network prune -f

# 6) Eliminar caches de builder
docker builder prune -af

echo "Docker cleanup completed."
