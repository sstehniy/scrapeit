#!/bin/sh
# wait-for-it.sh

set -e

host="$1"
shift
port="$1"
shift
cmd="$@"

until nc -z "$host" "$port"; do
  >&2 echo "Service $host:$port is unavailable - waiting..."
  sleep 1
done

>&2 echo "Service $host:$port is up - executing command"
exec $cmd
