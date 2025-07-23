#!/bin/sh
BASEDIR=$(dirname "$0")/../
CONFIG="configs/config-linux.yaml"
STOP_CMD="pkill rk-api"
START_CMD="$BASEDIR/rk-api web -c $CONFIG"
echo "stop landing server"
$STOP_CMD
echo "begin start landing server"
nohup $START_CMD  >/dev/null 2>&1 &
echo "finish start landing server"
