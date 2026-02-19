#!/bin/bash
export DISPLAY=:0
export DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/$(id -u)/bus

DIR=$(cd "$(dirname "$0")" && pwd)

OUTPUT=$("$DIR/registry-ping" -config "$DIR/config.yaml" 2>&1)

if [ -n "$OUTPUT" ]; then
    notify-send "registry-ping" "$OUTPUT"
fi
