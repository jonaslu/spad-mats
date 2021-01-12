#!/bin/bash
PG_DSN_STR=${PG_DSN:-postgres://postgres@localhost:5432/spad-mats?sslmode=disable}

HOST_USER_DB_REGEX="postgres://([^@]+)@([^/:]+):?([0-9]+)?/([^?]+)"
if [[ $PG_DSN_STR =~ $HOST_USER_DB_REGEX ]]; then
  USER="${BASH_REMATCH[1]}"
  HOST="${BASH_REMATCH[2]}"
  PORT="${BASH_REMATCH[3]}"
  DB="${BASH_REMATCH[4]}"
else
  echo "Malformed PG_DSN, could not capture host, user and database name"
  exit 1
fi

if [ "$PORT" = "" ]; then
  createdb -h $HOST -U $USER $DB
else
  createdb -h $HOST -U $USER -p $PORT $DB
fi

psql $PG_DSN_STR -f sql/create_table.sql $DB
