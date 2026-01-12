SELECT setval('crops_id_seq', (SELECT COALESCE(MAX(id), 0) FROM crops));
