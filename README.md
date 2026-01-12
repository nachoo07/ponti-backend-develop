# Ponti API

1. Ejecutar: make stg-build

2. Crear 2 crops, ejemplo:

curl --location 'localhost:8080/api/v1/crops/public/' \
--header 'Content-Type: application/json' \
--data '{
    "name":"wheat"
}'

3. Probar los endpoints de la coleccion Soalen.

