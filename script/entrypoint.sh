#!/bin/sh
set -e

if [ "${1:0:1}" = '-' ]; then
    set -- proxypool "$@"
fi

exec "$@"