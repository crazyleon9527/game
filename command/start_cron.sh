#!/bin/sh
# 设置基本目录和配置
BASEDIR=$(dirname "$0")/../
CONFIG="configs/config-linux.yaml"

# 进程检查命令
PROCESS_CHECK="pgrep -f rk-api"

# 停止命令
STOP_CMD="pkill rk-api"

# 启动命令
START_CMD="$BASEDIR/rk-api web -c $CONFIG"

# 检查进程是否存在
if ! $PROCESS_CHECK > /dev/null; then
    echo "rk-api process not found, attempting to restart."
    echo "Stopping server"
    $STOP_CMD  # 为了确保没有遗留进程，可选
    echo "Starting server"
    nohup $START_CMD > /dev/null 2>&1 &
    echo "rk-api service restarted"
else
    echo "rk-api service is running"
fi


   */5 * * * * /path/to/check_and_restart.sh >> /path/to/cron_log.log 2>&1