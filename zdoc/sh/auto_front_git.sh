#!/bin/bash

# 设置仓库的本地路径
REPO_PATH="/data/cheetah/cheetah-admin"

# 进入git仓库的根目录
cd $REPO_PATH

# 确保目前是在cip分支上
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "cip" ]; then
  echo "Switching to cip branch"
  git checkout cip
fi

# 拉取最新的代码变化，但不合并
git fetch

# 检测远端cip分支是否有更新
LOCAL=$(git rev-parse @)
REMOTE=$(git rev-parse @{u})

# 如果本地和远程不同步，就拉取(remote)最新代码并合并
if [ $LOCAL != $REMOTE ]; then
  echo "Updating cip branch"
  git pull origin cip
else
  echo "cip branch is up to date"
fi