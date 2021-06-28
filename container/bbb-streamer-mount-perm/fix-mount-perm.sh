#!/bin/sh

set -e

chown -R "$1" "$2"
chmod u=rwx "$2"

echo "Adjusted permissions for $1 on $2"
