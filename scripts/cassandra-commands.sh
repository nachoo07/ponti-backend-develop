#!/bin/bash

cqlsh
sql
USE cassandra_keyspace;
DESC TABLES;
DESC tweets;
SELECT * FROM tweets;

# limpiar tabla
TRUNCATE tweets;
TRUNCATE timeline_by_user;

SELECT * FROM timeline_by_user WHERE user_id = 'd89972a7-b476-407a-b98b-bd37d11a6e3e' ORDER BY created_at DESC;

SELECT * FROM timeline_by_user WHERE tweet_id = 'b062b1cf-12c1-46e1-96b4-29089facb2ac' ORDER BY created_at DESC;

