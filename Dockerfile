FROM postgres:latest
COPY initdb.sql /docker-entrypoint-initdb.d/
