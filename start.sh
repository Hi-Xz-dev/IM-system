#!/bin/sh
set -eu

mode=${1:-}
if [ "$#" -gt 1 ] || { [ "$#" -eq 1 ] && [ "$mode" != "--build" ] && [ "$mode" != "--preview" ]; }; then
    echo "用法：./start.sh [--build | --preview]" >&2
    exit 1
fi

cd "$(dirname "$0")"

pixel_style=''
ready_style=''
dim=''
reset=''
if [ -t 1 ] && [ -z "${NO_COLOR:-}" ] && [ "${TERM:-dumb}" != dumb ]; then
    pixel_style=$(printf '\033[48;5;')
    ready_style=$(printf '\033[1;30;102m')
    dim=$(printf '\033[2m')
    reset=$(printf '\033[0m')
fi

printf '\n'
if [ -n "$pixel_style" ]; then
    # 每个 # 对应一个上色格子。色阶沿列向右，并随行向下推进。
    awk -v pixel_style="$pixel_style" -v reset="$reset" '
        BEGIN {
            shades = split("63 69 75 81 87 123 122 121 120", palette)
        }
        {
            width = length($0)
            for (column = 1; column <= width; column++) {
                pixel = substr($0, column, 1)
                if (pixel == "#") {
                    shade = 1 + int((column - 1 + 2 * (NR - 1)) * shades / (width + 8))
                    printf "%s%dm %s", pixel_style, palette[shade], reset
                } else {
                    printf " "
                }
            }
            printf "\n"
        }
    ' <<'LOGO'
### #   #    ##### #   # ##### ##### ##### #   #
 #  ## ##    #      # #  #       #   #     ## ##
 #  # # #    #####   #   #####   #   ####  # # #
 #  #   #        #   #       #   #   #     #   #
### #   #    #####   #   #####   #   ##### #   #
LOGO
else
    printf 'IM System\n'
fi
printf '\n%sGo 即时通信%s\n' "$dim" "$reset"
printf '%s作者%s    Xz\n' "$dim" "$reset"
printf '%sGitHub%s  https://github.com/Hi-Xz-dev\n\n' "$dim" "$reset"

print_ready() {
    printf '\n%s ✓ IM 已就绪 %s  http://127.0.0.1:8081/web/\n' "$ready_style" "$reset"
    printf '%s在项目目录停止服务%s  docker compose down\n' "$dim" "$reset"
}

if [ "$mode" = "--preview" ]; then
    printf '%s预览模式 · 仅展示启动画面%s\n' "$dim" "$reset"
    print_ready
    exit 0
fi

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
print_ready
