FROM postgres:16
ENV POSTGRES_PASSWORD postgres
ENV POSTGRES_USER user
ENV POSTGRES_DB db

COPY initdb.sql /docker-entrypoint-initdb.d/
