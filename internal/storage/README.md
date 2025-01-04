
iot=# SELECT EXISTS (
SELECT FROM pg_tables
WHERE schemaname = 'public'
AND tablename = 'sensor'
);
 exists 
--------
 f
(1 row)
