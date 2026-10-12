#!/bin/bash
# =============================================================================
# check_numa_binding.sh
# 检查多个用户 home 目录下 .numanode 中的 CPU 绑定是否在用户之间重复
#
# 背景:
#   GoldenDB 部署时会把某个用户绑定的 CPU 核心写入该用户 home 目录下的
#   .numanode，内容形如：
#       0-23,48-71
#   表示该用户绑定了 0-23、48-71 两段核心。若不同用户绑定了同一个核心，
#   会造成进程互相争抢 CPU，需要在巡检阶段提前发现。
#
# 用法:
#   bash check_numa_binding.sh [选项]
#
# 选项:
#   --users <list>       目标用户列表，逗号或空格分隔（默认 user1,user2）
#   --user <name>        追加一个用户，可重复使用（与 --users 等价）
#   --home <user:dir>    指定某用户的 home 目录，可重复使用（默认从 passwd 解析）
#   --file <user:path>   指定某用户的 .numanode 路径，可重复使用（本地调试用）
#   --output <mode>      输出格式：json（默认，供主程序解析）/ text（人工查看）
#   --format <mode>      同 --output
#   --sudo               使用 sudo -n cat 读取 .numanode（home 无读权限时使用）
#   -h, --help           显示帮助
#
# 检查内容:
#   1. 同一用户的 .numanode 里是否重复列出同一个核心（如 0-7,4-9）
#   2. 同一个核心是否被多个用户重复绑定（核心问题，决定 status）
#   3. 绑定的核心是否真实存在于本机 NUMA 拓扑中，依据 lscpu 的
#      "NUMA nodeN CPU(s): 0-23" 行解析，并统计未被任何用户绑定的核心数
#
# 输出字段（--output json）:
#   status              ok        未发现跨用户重复绑定
#                       warning   存在无法检查的用户、超范围核心或用户内部重复
#                       duplicate 发现跨用户重复绑定的核心
#                       error     所有用户的 .numanode 都无法检查
#   info                结果说明（多条用；连接）
#   users               本次检查的用户列表，如 "user1,user2"
#   user_count          用户数
#   ok_user_count       检查成功的用户数
#   failed_user_count   检查失败的用户数
#   binding_count       各用户绑定核心条目数之和
#   cpu_count           所有用户去重后的核心数
#   repeat_count        同一用户内部重复列出的核心数合计
#   duplicate_count     被多个用户重复绑定的核心数
#   duplicate_cpus      重复核心列表，如 "0,1,2"
#   duplicates          重复明细 [{"cpu":0,"users":["user1","user2"],"count":2}]
#   out_of_range_count  不存在于本机 NUMA 拓扑中的核心数
#   out_of_range_cpus   超范围核心列表
#   unbound_count       本机未被任何用户绑定的核心数
#   numa_source         lscpu / unavailable
#   numa_node_count     NUMA 节点数
#   numa_nodes          NUMA 拓扑 [{"node":"0","cpus":"0-23","count":24}]
#   users_detail        用户明细 [{"user":...,"home":...,"file":...,"cpus":...,
#                                   "cpu_count":...,"repeat_count":...,"status":...,
#                                   "info":...}]
#
# 退出码:
#   0 检查完成（ok/warning）    1 脚本级错误或所有用户均检查失败    2 发现跨用户重复绑定
#
# 示例:
#   bash check_numa_binding.sh
#   bash check_numa_binding.sh --users user1,user2 --output text
#   bash check_numa_binding.sh --user user1 --file user1:/tmp/user1.numanode
# =============================================================================

set -uo pipefail

export PATH=/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:$PATH

SCRIPT_VERSION="2.0.0"
DEFAULT_USERS="user1,user2"

# ---------------------------------------------------------------------------
# 全局变量
# ---------------------------------------------------------------------------
USERS=()                                   # 目标用户
OVERRIDE_USERS=()                          # --home/--file 覆盖：用户
OVERRIDE_HOMES=()                          # --home 覆盖：目录
OVERRIDE_FILES=()                          # --file 覆盖：文件路径
OUTPUT_MODE="json"
USE_SUDO=0

WORK_DIR=""                                # 临时工作目录
PAIR_FILE=""                               # "核心 用户" 记录文件
HOST_FILE=""                               # 本机核心集合文件

NUMA_SOURCE="unavailable"
NODE_IDS=()                                # NUMA 节点编号
NODE_CPUS=()                               # NUMA 节点核心列表
NODE_COUNTS=()                             # NUMA 节点核心数
HOST_CORES_STR=""                          # 本机核心集合，形如 " 0 1 2 "
HOST_CORE_COUNT=0

U_NAME=()                                  # 以下 8 个数组按用户下标对齐
U_HOME=()
U_FILE=()
U_SPEC=()
U_COUNT=()
U_REPEAT=()
U_STATUS=()
U_INFO=()

OK_USERS=()
FAILED_USERS=()
DUP_LINES=()                               # "核心 user1,user2"
OUT_OF_RANGE=()                            # 超出本机拓扑的核心
INFO_PARTS=()

DUP_COUNT=0
DUP_CORES=""
OUT_OF_RANGE_COUNT=0
OUT_OF_RANGE_STR=""
UNBOUND_COUNT=0
BINDING_COUNT=0
CPU_COUNT=0
REPEAT_TOTAL=0
USER_COUNT=0
OK_USER_COUNT=0
FAILED_USER_COUNT=0
USERS_STR=""
STATUS=""
INFO=""

# ---------------------------------------------------------------------------
# 基础工具
# ---------------------------------------------------------------------------
usage() {
    sed -n '/^# 用法:/,/^# =====/p' "$0" | sed 's/^# \{0,1\}//'
    exit 0
}

die() {
    echo "[错误] $*" >&2
    exit 1
}

to_lower() {
    printf '%s' "${1-}" | tr 'A-Z' 'a-z'
}

json_escape() {
    local s="${1-}"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/\\r}"
    s="${s//$'\n'/\\n}"
    printf '%s' "$s"
}

# 数组是否包含指定元素（数组可能为空，需避免 bash 3.2 下 set -u 报错）
array_contains() {
    local needle="$1"
    shift
    (( $# > 0 )) || return 1
    local item
    for item in "$@"; do
        [[ "$item" == "$needle" ]] && return 0
    done
    return 1
}

# 清理临时目录
cleanup() {
    if [[ -n "$WORK_DIR" && -d "$WORK_DIR" ]]; then
        rm -rf "$WORK_DIR"
    fi
    return 0
}

# ---------------------------------------------------------------------------
# 用户参数处理
# ---------------------------------------------------------------------------
# 把 "user1,user2 user3" 形式的列表追加到 USERS（自动去重）
add_users() {
    local spec="${1//,/ }"
    local item
    for item in $spec; do
        [[ -n "$item" ]] || continue
        [[ "$item" =~ ^[A-Za-z0-9._-]+$ ]] || die "非法的用户名: ${item}"
        if array_contains "$item" ${USERS[@]+"${USERS[@]}"}; then
            continue
        fi
        USERS+=("$item")
    done
}

# 记录 <user>:<path> 形式的覆盖项，path 允许包含冒号
set_override() {
    local pair="$1"
    local kind="$2"
    local user="${pair%%:*}"
    local path="${pair#*:}"
    [[ -n "$user" && -n "$path" ]] || die "覆盖项格式错误（应为 <user>:<path>）: ${pair}"

    OVERRIDE_USERS+=("$user")
    if [[ "$kind" == "home" ]]; then
        OVERRIDE_HOMES+=("$path")
        OVERRIDE_FILES+=("")
    else
        OVERRIDE_HOMES+=("")
        OVERRIDE_FILES+=("$path")
    fi
}

# 查询某个用户的覆盖值，取最后一次设置；没有则输出空串
override_for() {
    local user="$1"
    local kind="$2"
    local i result=""
    for ((i = 0; i < ${#OVERRIDE_USERS[@]}; i++)); do
        [[ "${OVERRIDE_USERS[$i]}" == "$user" ]] || continue
        if [[ "$kind" == "home" ]]; then
            result="${OVERRIDE_HOMES[$i]}"
        else
            result="${OVERRIDE_FILES[$i]}"
        fi
    done
    printf '%s' "$result"
}

# 解析用户 home 目录：getent 优先，其次 /etc/passwd，最后按约定回退
resolve_home() {
    local user="$1" home=""
    if command -v getent >/dev/null 2>&1; then
        home="$(getent passwd "$user" 2>/dev/null | awk -F: 'NR == 1 {print $6}')"
    fi
    if [[ -z "$home" && -r /etc/passwd ]]; then
        home="$(awk -F: -v u="$user" '$1 == u {print $6; exit}' /etc/passwd)"
    fi
    if [[ -z "$home" ]]; then
        if [[ "$user" == "root" ]]; then
            home="/root"
        else
            home="/home/${user}"
        fi
    fi
    printf '%s' "$home"
}

# ---------------------------------------------------------------------------
# CPU 表达式处理
# ---------------------------------------------------------------------------
# 把 .numanode 内容整理成单行 CPU 表达式：去掉 # 注释与 CR，多行合并为逗号
clean_spec() {
    printf '%s\n' "${1-}" \
        | sed 's/#.*$//' \
        | tr -d '\r' \
        | tr '\n' ',' \
        | sed 's/[[:space:]]*,[[:space:]]*/,/g; s/,\{1,\}/,/g; s/^,//; s/,$//'
}

# 展开 CPU 表达式，每行输出一个核心号，支持：
#   0-23        区间
#   0,1,2       离散核心
#   0-7 16-23   空格分隔的多段
#   cpu0-cpu3   带 cpu/core 前缀
# 表达式非法时向 stderr 输出原因并返回 1
expand_cpu_list() {
    local spec="${1-}"
    local normalized token start end core

    normalized="$(printf '%s' "$spec" \
        | tr ',\t;|' '    ' \
        | sed 's/[[:space:]]*-[[:space:]]*/-/g')"

    for token in $normalized; do
        token="${token//\"/}"
        token="${token//\'/}"
        start="${token%%-*}"
        end="${token##*-}"
        start="${start#cpu}"
        start="${start#core}"
        end="${end#cpu}"
        end="${end#core}"
        if [[ ! "$start" =~ ^[0-9]+$ || ! "$end" =~ ^[0-9]+$ ]]; then
            printf '无法识别的 CPU 表达式: %s\n' "$token" >&2
            return 1
        fi
        if (( start > end )); then
            printf '区间起止顺序错误: %s\n' "$token" >&2
            return 1
        fi
        for ((core = start; core <= end; core++)); do
            printf '%s\n' "$core"
        done
    done
    return 0
}

# 统计 CPU 表达式去重后的核心数，解析失败按 0 处理
count_cores() {
    local spec="${1-}"
    expand_cpu_list "$spec" 2>/dev/null | LC_ALL=C sort -u | grep -c . || true
}

# 按数值升序排序并去重，用逗号连接（用于输出核心列表）
join_cores() {
    LC_ALL=C sort -n -u | paste -sd, -
}

# ---------------------------------------------------------------------------
# lscpu 拓扑
# ---------------------------------------------------------------------------
# 解析 lscpu 的 "NUMA node0 CPU(s): 0-23" 行，填充 NODE_* 数组
detect_numa_topo() {
    local lines line node spec
    command -v lscpu >/dev/null 2>&1 || return 0

    lines="$(LC_ALL=C lscpu 2>/dev/null | grep -E '^NUMA node[0-9]+ CPU\(s\):' || true)"
    [[ -n "$lines" ]] || return 0

    while IFS= read -r line; do
        node="${line#NUMA node}"
        node="${node%% *}"
        spec="${line##*: }"
        [[ -n "$node" && -n "$spec" ]] || continue
        NODE_IDS+=("$node")
        NODE_CPUS+=("$spec")
        NODE_COUNTS+=("$(count_cores "$spec")")
        NUMA_SOURCE="lscpu"
    done <<< "$lines"
}

# 生成本机核心集合，用于范围校验与未绑定统计
build_host_cores() {
    local i
    (( ${#NODE_CPUS[@]} > 0 )) || return 0

    for ((i = 0; i < ${#NODE_CPUS[@]}; i++)); do
        expand_cpu_list "${NODE_CPUS[$i]}" 2>/dev/null || true
    done | LC_ALL=C sort -u > "$HOST_FILE"

    HOST_CORE_COUNT="$(grep -c . "$HOST_FILE" || true)"
    if (( HOST_CORE_COUNT > 0 )); then
        HOST_CORES_STR=" $(paste -sd' ' "$HOST_FILE") "
    fi
}

# ---------------------------------------------------------------------------
# 单个用户结果记录
# ---------------------------------------------------------------------------
record_user_failure() {
    local user="$1" home="$2" file="$3" info="$4"
    FAILED_USERS+=("$user")
    U_NAME+=("$user")
    U_HOME+=("$home")
    U_FILE+=("$file")
    U_SPEC+=("")
    U_COUNT+=(0)
    U_REPEAT+=(0)
    U_STATUS+=("error")
    U_INFO+=("$info")
}

record_user_success() {
    local user="$1" home="$2" file="$3" spec="$4" count="$5" repeat="$6"
    OK_USERS+=("$user")
    U_NAME+=("$user")
    U_HOME+=("$home")
    U_FILE+=("$file")
    U_SPEC+=("$spec")
    U_COUNT+=("$count")
    U_REPEAT+=("$repeat")
    U_STATUS+=("ok")
    U_INFO+=("")
}

# ---------------------------------------------------------------------------
# 参数解析
# ---------------------------------------------------------------------------
ARGS=()
for a in "$@"; do
    case "$a" in
        --*=*) ARGS+=("${a%%=*}" "${a#*=}") ;;
        *)     ARGS+=("$a") ;;
    esac
done
if (( ${#ARGS[@]} > 0 )); then
    set -- "${ARGS[@]}"
else
    set --
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            usage
            ;;
        --users|--user)
            [[ $# -ge 2 ]] || die "$1 缺少参数"
            add_users "$2"
            shift 2
            ;;
        --home|--file)
            [[ $# -ge 2 ]] || die "$1 缺少参数"
            [[ "$2" == *:* ]] || die "$1 需要 <user>:<path> 形式，当前值: $2"
            if [[ "$1" == "--home" ]]; then
                set_override "$2" home
            else
                set_override "$2" file
            fi
            shift 2
            ;;
        --output|--format)
            [[ $# -ge 2 ]] || die "$1 缺少参数"
            OUTPUT_MODE="$(to_lower "$2")"
            shift 2
            ;;
        --sudo)
            USE_SUDO=1
            shift
            ;;
        *)
            die "未知参数: $1（可用 --help 查看用法）"
            ;;
    esac
done

case "$OUTPUT_MODE" in
    json|text) ;;
    *) die "--output 仅支持 json 或 text，当前值: ${OUTPUT_MODE}" ;;
esac

if (( ${#USERS[@]} == 0 )); then
    add_users "$DEFAULT_USERS"
fi

# ---------------------------------------------------------------------------
# 临时工作目录
# ---------------------------------------------------------------------------
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/.numa_check.XXXXXX" 2>/dev/null)" \
    || die "无法创建临时目录"
PAIR_FILE="$WORK_DIR/pairs.txt"
HOST_FILE="$WORK_DIR/host_cores.txt"
: > "$PAIR_FILE"
: > "$HOST_FILE"
trap cleanup EXIT

# ---------------------------------------------------------------------------
# 步骤 1：本机 NUMA 拓扑
# ---------------------------------------------------------------------------
detect_numa_topo
build_host_cores

# ---------------------------------------------------------------------------
# 步骤 2：逐用户读取并展开 .numanode
# ---------------------------------------------------------------------------
for user in ${USERS[@]+"${USERS[@]}"}; do
    home="$(resolve_home "$user")"
    override="$(override_for "$user" home)"
    [[ -n "$override" ]] && home="$override"

    file="${home%/}/.numanode"
    override="$(override_for "$user" file)"
    [[ -n "$override" ]] && file="$override"

    if (( USE_SUDO == 1 )); then
        if ! content="$(sudo -n cat "$file" 2>"$WORK_DIR/read.err")"; then
            record_user_failure "$user" "$home" "$file" \
                "sudo 读取失败: $(head -n 1 "$WORK_DIR/read.err")"
            continue
        fi
    else
        if [[ ! -e "$file" ]]; then
            record_user_failure "$user" "$home" "$file" "文件不存在: ${file}"
            continue
        fi
        if [[ ! -f "$file" ]]; then
            record_user_failure "$user" "$home" "$file" "不是普通文件: ${file}"
            continue
        fi
        if [[ ! -r "$file" ]]; then
            record_user_failure "$user" "$home" "$file" "文件不可读（可尝试 --sudo）"
            continue
        fi
        if ! content="$(cat "$file" 2>"$WORK_DIR/read.err")"; then
            record_user_failure "$user" "$home" "$file" \
                "读取失败: $(head -n 1 "$WORK_DIR/read.err")"
            continue
        fi
    fi

    spec="$(clean_spec "$content")"
    if [[ -z "$spec" ]]; then
        record_user_failure "$user" "$home" "$file" "文件内容为空"
        continue
    fi

    if ! cores="$(expand_cpu_list "$spec" 2>"$WORK_DIR/expand.err")"; then
        record_user_failure "$user" "$home" "$file" \
            "解析失败: $(head -n 1 "$WORK_DIR/expand.err")"
        continue
    fi

    unique_cores="$(printf '%s\n' "$cores" | LC_ALL=C sort -u)"
    raw_count="$(printf '%s\n' "$cores" | grep -c . || true)"
    unique_count="$(printf '%s\n' "$unique_cores" | grep -c . || true)"
    repeat=$(( raw_count - unique_count ))
    REPEAT_TOTAL=$(( REPEAT_TOTAL + repeat ))

    # 记录 "核心 用户"，供跨用户重复检测使用
    printf '%s\n' "$unique_cores" | sed "s/\$/ ${user}/" >> "$PAIR_FILE"

    record_user_success "$user" "$home" "$file" "$spec" "$unique_count" "$repeat"
done

# ---------------------------------------------------------------------------
# 步骤 3：跨用户重复检测
# ---------------------------------------------------------------------------
LC_ALL=C sort -k1,1 -k2,2 -u "$PAIR_FILE" > "$WORK_DIR/pairs.sorted"

BINDING_COUNT="$(cut -d' ' -f2 "$WORK_DIR/pairs.sorted" | grep -c . || true)"
CPU_COUNT="$(cut -d' ' -f1 "$WORK_DIR/pairs.sorted" | uniq | grep -c . || true)"

prev_core=""
owners=""
owners_count=0
while read -r core owner; do
    [[ -n "$core" ]] || continue
    if [[ "$core" == "$prev_core" ]]; then
        owners="${owners},${owner}"
        owners_count=$(( owners_count + 1 ))
    else
        if (( owners_count > 1 )); then
            DUP_LINES+=("${prev_core} ${owners}")
        fi
        prev_core="$core"
        owners="$owner"
        owners_count=1
    fi
done < "$WORK_DIR/pairs.sorted"
if (( owners_count > 1 )); then
    DUP_LINES+=("${prev_core} ${owners}")
fi

DUP_COUNT="${#DUP_LINES[@]}"
if (( DUP_COUNT > 0 )); then
    DUP_CORES="$(printf '%s\n' "${DUP_LINES[@]}" | cut -d' ' -f1 | join_cores)"
fi

# ---------------------------------------------------------------------------
# 步骤 4：与 lscpu 拓扑比对
# ---------------------------------------------------------------------------
if [[ -n "$HOST_CORES_STR" ]]; then
    cut -d' ' -f1 "$WORK_DIR/pairs.sorted" | uniq > "$WORK_DIR/bound_cores.txt"
    bound_in_host=0
    while IFS= read -r core; do
        [[ -n "$core" ]] || continue
        if [[ "$HOST_CORES_STR" == *" $core "* ]]; then
            bound_in_host=$(( bound_in_host + 1 ))
        else
            OUT_OF_RANGE+=("$core")
        fi
    done < "$WORK_DIR/bound_cores.txt"

    OUT_OF_RANGE_COUNT="${#OUT_OF_RANGE[@]}"
    if (( OUT_OF_RANGE_COUNT > 0 )); then
        OUT_OF_RANGE_STR="$(printf '%s\n' "${OUT_OF_RANGE[@]}" | join_cores)"
    fi
    UNBOUND_COUNT=$(( HOST_CORE_COUNT - bound_in_host ))
fi

# ---------------------------------------------------------------------------
# 步骤 5：结论
# ---------------------------------------------------------------------------
OK_USER_COUNT="${#OK_USERS[@]}"
FAILED_USER_COUNT="${#FAILED_USERS[@]}"
USER_COUNT="${#USERS[@]}"
USERS_STR="$(IFS=','; printf '%s' "${USERS[*]}")"

if (( DUP_COUNT > 0 )); then
    STATUS="duplicate"
elif (( OK_USER_COUNT == 0 )); then
    STATUS="error"
elif (( FAILED_USER_COUNT > 0 || OUT_OF_RANGE_COUNT > 0 || REPEAT_TOTAL > 0 )); then
    STATUS="warning"
else
    STATUS="ok"
fi

if (( DUP_COUNT > 0 )); then
    INFO_PARTS+=("检测到 ${DUP_COUNT} 个 CPU 核心被多个用户重复绑定: ${DUP_CORES}")
fi
if (( FAILED_USER_COUNT > 0 )); then
    INFO_PARTS+=("${FAILED_USER_COUNT}/${USER_COUNT} 个用户的 .numanode 检查失败")
fi
if (( OUT_OF_RANGE_COUNT > 0 )); then
    INFO_PARTS+=("${OUT_OF_RANGE_COUNT} 个核心不在本机 NUMA 拓扑范围内: ${OUT_OF_RANGE_STR}")
fi
if (( REPEAT_TOTAL > 0 )); then
    INFO_PARTS+=("${REPEAT_TOTAL} 个核心在同一用户的绑定列表中被重复列出")
fi
if (( ${#INFO_PARTS[@]} == 0 )); then
    INFO_PARTS=("未发现多个用户之间重复绑定的 CPU 核心（用户 ${USER_COUNT} 个，去重核心 ${CPU_COUNT} 个）")
fi

for part in ${INFO_PARTS[@]+"${INFO_PARTS[@]}"}; do
    if [[ -n "$INFO" ]]; then
        INFO="${INFO}；${part}"
    else
        INFO="$part"
    fi
done

# ---------------------------------------------------------------------------
# 输出：JSON
# ---------------------------------------------------------------------------
emit_json() {
    local i j
    local duplicates_json="" nodes_json="" details_json="" owners_json=""
    local owners=()

    for ((i = 0; i < ${#DUP_LINES[@]}; i++)); do
        local core="${DUP_LINES[$i]%% *}"
        local owner_list="${DUP_LINES[$i]#* }"
        IFS=',' read -r -a owners <<< "$owner_list"
        owners_json=""
        for ((j = 0; j < ${#owners[@]}; j++)); do
            [[ $j -gt 0 ]] && owners_json+=","
            owners_json+="\"$(json_escape "${owners[$j]}")\""
        done
        [[ $i -gt 0 ]] && duplicates_json+=","
        duplicates_json+="{\"cpu\":${core},\"users\":[${owners_json}],\"count\":${#owners[@]}}"
    done

    for ((i = 0; i < ${#NODE_IDS[@]}; i++)); do
        [[ $i -gt 0 ]] && nodes_json+=","
        nodes_json+="{\"node\":\"$(json_escape "${NODE_IDS[$i]}")\","
        nodes_json+="\"cpus\":\"$(json_escape "${NODE_CPUS[$i]}")\","
        nodes_json+="\"count\":${NODE_COUNTS[$i]}}"
    done

    for ((i = 0; i < ${#U_NAME[@]}; i++)); do
        [[ $i -gt 0 ]] && details_json+=","
        details_json+="{\"user\":\"$(json_escape "${U_NAME[$i]}")\","
        details_json+="\"home\":\"$(json_escape "${U_HOME[$i]}")\","
        details_json+="\"file\":\"$(json_escape "${U_FILE[$i]}")\","
        details_json+="\"cpus\":\"$(json_escape "${U_SPEC[$i]}")\","
        details_json+="\"cpu_count\":${U_COUNT[$i]},"
        details_json+="\"repeat_count\":${U_REPEAT[$i]},"
        details_json+="\"status\":\"${U_STATUS[$i]}\","
        details_json+="\"info\":\"$(json_escape "${U_INFO[$i]}")\"}"
    done

    printf '{"status":"%s","info":"%s",' \
        "$STATUS" "$(json_escape "$INFO")"
    printf '"users":"%s","user_count":%d,"ok_user_count":%d,"failed_user_count":%d,' \
        "$(json_escape "$USERS_STR")" "$USER_COUNT" "$OK_USER_COUNT" "$FAILED_USER_COUNT"
    printf '"binding_count":%d,"cpu_count":%d,"repeat_count":%d,' \
        "$BINDING_COUNT" "$CPU_COUNT" "$REPEAT_TOTAL"
    printf '"duplicate_count":%d,"duplicate_cpus":"%s","duplicates":[%s],' \
        "$DUP_COUNT" "$(json_escape "$DUP_CORES")" "$duplicates_json"
    printf '"out_of_range_count":%d,"out_of_range_cpus":"%s","unbound_count":%d,' \
        "$OUT_OF_RANGE_COUNT" "$(json_escape "$OUT_OF_RANGE_STR")" "$UNBOUND_COUNT"
    printf '"numa_source":"%s","numa_node_count":%d,"numa_nodes":[%s],' \
        "$(json_escape "$NUMA_SOURCE")" "${#NODE_IDS[@]}" "$nodes_json"
    printf '"users_detail":[%s]}\n' "$details_json"
}

# ---------------------------------------------------------------------------
# 输出：文本
# ---------------------------------------------------------------------------
emit_text() {
    local i

    echo "NUMA 绑定重复检查（多用户）"
    echo "  脚本版本 : ${SCRIPT_VERSION}"
    echo "  用户列表 : ${USERS_STR}"
    echo ""

    if (( ${#NODE_IDS[@]} > 0 )); then
        echo "本机 NUMA 拓扑（lscpu）:"
        for ((i = 0; i < ${#NODE_IDS[@]}; i++)); do
            echo "  node${NODE_IDS[$i]} | ${NODE_CPUS[$i]} | ${NODE_COUNTS[$i]} 个核心"
        done
        echo "  本机核心合计 ${HOST_CORE_COUNT} 个 | 已绑定 ${CPU_COUNT} 个 | 未绑定 ${UNBOUND_COUNT} 个"
    else
        echo "本机 NUMA 拓扑（lscpu）: 未获取到 NUMA 信息，跳过范围校验"
    fi
    echo ""

    echo "用户绑定明细:"
    for ((i = 0; i < ${#U_NAME[@]}; i++)); do
        if [[ "${U_STATUS[$i]}" == "ok" ]]; then
            printf '  [%d] %s | %s | %s | %s 个核心' \
                "$((i + 1))" "${U_NAME[$i]}" "${U_FILE[$i]}" "${U_SPEC[$i]}" "${U_COUNT[$i]}"
            if (( ${U_REPEAT[$i]} > 0 )); then
                printf '（内部重复 %s 个）' "${U_REPEAT[$i]}"
            fi
            echo ""
        else
            printf '  [%d] %s | %s | [错误] %s\n' \
                "$((i + 1))" "${U_NAME[$i]}" "${U_FILE[$i]}" "${U_INFO[$i]}"
        fi
    done
    echo ""

    echo "汇总: 用户 ${USER_COUNT} 个（成功 ${OK_USER_COUNT} / 失败 ${FAILED_USER_COUNT}）| 绑定条目 ${BINDING_COUNT} | 去重核心 ${CPU_COUNT} | 重复核心 ${DUP_COUNT} | 超范围核心 ${OUT_OF_RANGE_COUNT} | 未绑定核心 ${UNBOUND_COUNT}"

    if (( DUP_COUNT > 0 )); then
        echo ""
        echo "跨用户重复核心:"
        for ((i = 0; i < ${#DUP_LINES[@]}; i++)); do
            printf '  CPU %s | 绑定用户: %s\n' "${DUP_LINES[$i]%% *}" "${DUP_LINES[$i]#* }"
        done
    fi

    echo ""
    case "$STATUS" in
        ok)        echo "结果: [OK] ${INFO}" ;;
        warning)   echo "结果: [警告] ${INFO}" ;;
        duplicate) echo "结果: [重复] ${INFO}" ;;
        *)         echo "结果: [错误] ${INFO}" ;;
    esac
}

if [[ "$OUTPUT_MODE" == "json" ]]; then
    emit_json
else
    emit_text
fi

case "$STATUS" in
    duplicate) exit 2 ;;
    error)     exit 1 ;;
    *)         exit 0 ;;
esac
