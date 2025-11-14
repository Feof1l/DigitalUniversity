FROM postgres:16-alpine

ENV POSTGRES_PASSWORD=postgres \
    POSTGRES_USER=user \
    POSTGRES_DB=db

COPY initdb.sql /docker-entrypoint-initdb.d/
