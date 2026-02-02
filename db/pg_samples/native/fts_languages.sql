-- All available language
SELECT cfgname
FROM pg_ts_config;

-- More detailed view with schema
SELECT nspname AS schema, cfgname AS configuration
FROM pg_ts_config
         JOIN pg_namespace ON pg_ts_config.cfgnamespace = pg_namespace.oid
ORDER BY nspname, cfgname;

-- Even more details (including comments/descriptions)
SELECT
    nspname AS schema,
    cfgname AS configuration,
    obj_description(pg_ts_config.oid) AS description
FROM pg_ts_config
         JOIN pg_namespace ON pg_ts_config.cfgnamespace = pg_namespace.oid
ORDER BY nspname, cfgname;