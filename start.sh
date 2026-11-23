#!/bin/sh
set -eu

if [ "$#" -gt 1 ] || { [ "$#" -eq 1 ] && [ "$1" != "--build" ]; }; then
    echo "用法：./start.sh [--build]" >&2
    exit 1
fi

cd "$(dirname "$0")"

if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
    echo "请先安装 Docker 和 Docker Compose。" >&2
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo "正在启动 Docker Desktop…"
    if ! docker desktop start --timeout 90 || ! docker info >/dev/null 2>&1; then
        echo "无法连接 Docker，请启动 Docker Desktop 或 Docker Engine 后重试。" >&2
        exit 1
    fi
fi

if [ ! -e .env ]; then
    umask 077
    env_tmp=$(mktemp .env.tmp.XXXXXX)
    trap 'rm -f "$env_tmp"' EXIT
    for key in JWT_SECRET MYSQL_ROOT_PASSWORD MYSQL_PASSWORD RABBITMQ_PASSWORD; do
        value=$(od -An -N32 -tx1 /dev/urandom | tr -d '[:space:]')
        if [ "${#value}" -ne 64 ]; then
            echo "生成随机密钥失败。" >&2
            exit 1
        fi
        printf '%s=%s\n' "$key" "$value" >> "$env_tmp"
    done
    # 原子创建，重复运行时沿用此前写入的密钥。
    if ln "$env_tmp" .env 2>/dev/null; then
        echo "已生成本地配置 .env。"
    elif [ ! -f .env ]; then
        echo "无法创建 .env。" >&2
        exit 1
    fi
fi

docker compose up -d --wait --wait-timeout 180 "$@"
printf '\nIM 已就绪：http://127.0.0.1:8081/web/\n'
