#!/bin/bash

redis-cli -a "defaultpassword"
FLUSHALL # borra absolutamente todo en Redis.  
FLUSHDB # solo limpia la base de datos actual.  
KEYS * # Para verificar que Redis está vacío 
SELECT <número_db> # Para verificar en qué base de datos estás, ejemplo "0"