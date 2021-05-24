#!/bin/bash

set -e
set -u

if [ -n "$POSTGRES_DBS" ]; then
  dbs=$(echo "$POSTGRES_DBS" | tr "," " ")
  for db in $dbs
  do
    echo "Creating database $db"
    createdb -U "$POSTGRES_USER" -O "$POSTGRES_USER" "$db"
  done
fi
