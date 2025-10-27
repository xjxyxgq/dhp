# CMSys Mock 接口使用说明

## 概述

本文档介绍用于开发测试的 CMSys HTTP 接口 Mock 实现。Mock 接口模拟了真实 CMSys 系统的认证和数据查询功能，使得开发人员无需连接实际的 CMSys 系统即可进行开发和测试。

## Mock 接口列表

### 1. CMSys 认证接口（Mock）

**接口地址**
```
POST http://localhost:8888/platform/cmsys/auth
```

**请求示例**
```bash
curl -X POST 'http://localhost:8888/platform/cmsys/auth' \
  -H 'Content-Type: application/json' \
  -d '{
    "appCode": "DB",
    "secret": "your-app-secret"
  }'
```

**响应示例**
```json
{
  "code": "A0000",
  "msg": "success",
  "data": "mock-cmsys-token-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
}
```

**特点**
- 接受任意 appCode 和 secret 组合
- 返回固定的 mock token，方便测试
- 无需真实的认证逻辑

### 2. CMSys 数据接口（Mock）

**接口地址**
```
GET http://localhost:8888/platform/cmsys/data
```

**请求头**
- `x-control-access-token`: 认证 token（从 auth 接口获取）
- `x-control-access-operator`: 操作员标识（如 "admin"）

**请求参数**
- `query` (可选): 查询参数字符串

**请求示例**
```bash
curl -X GET 'http://localhost:8888/platform/cmsys/data?query=department=DB' \
  -H 'x-control-access-token: mock-cmsys-token-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9' \
  -H 'x-control-access-operator: admin'
```

**响应示例**
```json
{
  "code": "A000",
  "msg": "success",
  "data": [
    {
      "ipAddress": "192.168.1.1",
      "cpuMaxNew": "75.32",
      "memMaxNew": "68.45",
      "diskMaxNew": "55.89",
      "remark": "生产环境数据库服务器"
    },
    {
      "ipAddress": "192.168.1.2",
      "cpuMaxNew": "62.15",
      "memMaxNew": "72.34",
      "diskMaxNew": "48.67",
      "remark": "测试环境应用服务器"
    }
  ]
}
```

**特点**
- 自动生成 10 台主机的模拟数据
- 每次请求生成随机的资源利用率数值（40-90%）
- 包含真实场景的 remark 描述
- 验证 token 是否存在（若未提供则返回错误）

### 3. CMSys 基于IP列表查询数据接口（Mock）

**接口地址**
```
POST http://localhost:8888/platform/cmsys/data-by-ips
```

**请求头**
- `x-control-access-token`: 认证 token（从 auth 接口获取）
- `x-control-access-operator`: 操作员标识（如 "admin"）

**请求体**
```json
{
  "ips": ["192.168.1.1", "192.168.1.2", "10.0.0.100"]
}
```

**请求示例**
```bash
curl -X POST 'http://localhost:8888/platform/cmsys/data-by-ips' \
  -H 'Content-Type: application/json' \
  -H 'x-control-access-token: mock-cmsys-token-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9' \
  -H 'x-control-access-operator: admin' \
  -d '{
    "ips": ["192.168.1.1", "192.168.1.2", "10.0.0.100"]
  }'
```

**响应示例**
```json
{
  "code": "A000",
  "msg": "success",
  "data": [
    {
      "ipAddress": "192.168.1.1",
      "cpuMaxNew": "75.32",
      "memMaxNew": "68.45",
      "diskMaxNew": "55.89",
      "remark": "生产环境数据库服务器"
    },
    {
      "ipAddress": "192.168.1.2",
      "cpuMaxNew": "62.15",
      "memMaxNew": "72.34",
      "diskMaxNew": "48.67",
      "remark": "测试环境应用服务器"
    },
    {
      "ipAddress": "10.0.0.100",
      "cpuMaxNew": "45.78",
      "memMaxNew": "82.11",
      "diskMaxNew": "67.33",
      "remark": "开发环境Web服务器"
    }
  ]
}
```

**特点**
- 根据提供的 IP 列表生成对应数量的模拟数据
- 每个 IP 生成独立的随机资源利用率数值（40-90%）
- 支持任意数量的 IP 查询
- 验证 token 和 IP 列表参数
- IP 列表不能为空，否则返回错误

**错误响应示例**

未提供token:
```json
{
  "code": "A0001",
  "msg": "未提供认证token",
  "data": null
}
```

IP列表为空:
```json
{
  "code": "A0003",
  "msg": "IP列表不能为空",
  "data": null
}
```

## 配置说明

### 使用 Mock 接口（开发测试）

在 `rpc/etc/cmpool.yaml` 中配置：

```yaml
CMSysDataSource:
  AuthEndpoint: "http://localhost:8888/platform/cmsys/auth"
  DataEndpoint: "http://localhost:8888/platform/cmsys/data"
  AppCode: "DB"
  AppSecret: "your-app-secret-here"
  Operator: "admin"
  TimeoutSeconds: 60
```

### 切换到真实接口（生产环境）

```yaml
CMSysDataSource:
  AuthEndpoint: "https://api.cmsys.example.com/auth"
  DataEndpoint: "https://api.cmsys.example.com/data"
  AppCode: "DB"
  AppSecret: "your-real-app-secret"
  Operator: "admin"
  TimeoutSeconds: 60
```

## 完整测试流程

### 1. 启动服务

```bash
# 在项目根目录执行
./start.sh
```

或手动启动：

```bash
# 终端1：启动 RPC 服务
cd cmdb_backend_v2/rpc
go run cmpool.go -f etc/cmpool.yaml

# 终端2：启动 API 服务
cd cmdb_backend_v2/api
go run cmdb.go -f etc/cmdb-api.yaml
```

### 2. 测试认证接口

```bash
# 获取 token
curl -X POST 'http://localhost:8888/platform/cmsys/auth' \
  -H 'Content-Type: application/json' \
  -d '{
    "appCode": "DB",
    "secret": "test-secret"
  }'
```

### 3. 测试数据接口

```bash
# 使用获取的 token 查询数据（按组查询）
curl -X GET 'http://localhost:8888/platform/cmsys/data' \
  -H 'x-control-access-token: mock-cmsys-token-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9' \
  -H 'x-control-access-operator: admin'
```

### 4. 测试基于IP列表查询接口

```bash
# 使用获取的 token 查询指定IP的数据
curl -X POST 'http://localhost:8888/platform/cmsys/data-by-ips' \
  -H 'Content-Type: application/json' \
  -H 'x-control-access-token: mock-cmsys-token-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9' \
  -H 'x-control-access-operator: admin' \
  -d '{
    "ips": ["192.168.1.1", "192.168.1.2", "10.0.0.100"]
  }'
```

### 5. 测试完整同步流程

```bash
# 执行 CMSys 数据同步
curl -X POST 'http://localhost:8888/api/cmdb/v1/cmsys-sync' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_AUTH_TOKEN' \
  -d '{
    "task_name": "Mock接口测试同步"
  }'
```

## Mock 数据说明

### 生成的主机数据

Mock 接口会生成 10 台虚拟主机的数据：

| IP 地址 | 备注说明 |
|---------|----------|
| 192.168.1.1 | 生产环境数据库服务器 |
| 192.168.1.2 | 测试环境应用服务器 |
| 192.168.1.3 | 开发环境Web服务器 |
| 192.168.1.4 | 备份服务器 |
| 192.168.1.5 | 监控服务器 |
| 192.168.1.6 | 日志服务器 |
| 192.168.1.7 | 缓存服务器 |
| 192.168.1.8 | 队列服务器 |
| 192.168.1.9 | 存储服务器 |
| 192.168.1.10 | 负载均衡服务器 |

### 资源利用率范围

- **CPU 利用率**: 40% - 90%
- **内存利用率**: 50% - 90%
- **磁盘利用率**: 30% - 90%

所有数值都是随机生成的，每次请求都会不同。

## 实现代码位置

### API 层

- **认证 Handler**: `api/internal/handler/mockcmsysauthhandler.go`
- **数据 Handler**: `api/internal/handler/mockcmsysdatahandler.go`

### API 定义

- **路由定义**: `api/cmdb.api` (第 1120-1124 行)

## 与 ES Mock 接口的对比

| 特性 | CMSys Mock (按组) | CMSys Mock (按IP) | ES Mock |
|------|------------------|-------------------|---------|
| 认证方式 | 需要 token 认证 | 需要 token 认证 | 无需认证 |
| 请求方式 | GET + Header | POST + Body | POST + Body |
| 查询方式 | 按组查询(默认10台) | 按IP列表查询 | 按集群/IP查询 |
| 数据格式 | 汇总数据 | 汇总数据 | 时序聚合数据 |
| 数据字段 | ipAddress, cpuMaxNew, memMaxNew, diskMaxNew, remark | ipAddress, cpuMaxNew, memMaxNew, diskMaxNew, remark | 聚合统计 (max, avg, count) |
| 主机数量 | 固定 10 台 | 根据IP列表动态生成 | 动态生成 |
| 使用场景 | 批量查询组内主机 | 精确查询指定IP | ES数据源测试 |

## 开发建议

1. **开发阶段**: 使用 Mock 接口，快速迭代开发
2. **集成测试**: 先用 Mock 接口验证逻辑正确性
3. **联调阶段**: 切换到真实接口，验证数据格式兼容性
4. **生产部署**: 使用真实 CMSys 接口地址

## 故障排查

### 问题 1: 认证接口返回 404

**原因**: API 服务未启动或端口配置错误

**解决**:
```bash
# 检查 API 服务是否运行
ps aux | grep cmdb-api

# 检查端口是否监听
lsof -i :8888
```

### 问题 2: 数据接口返回"未提供认证token"

**原因**: 请求头中未包含 `x-control-access-token`

**解决**: 确保在请求头中添加 token
```bash
curl -H 'x-control-access-token: YOUR_TOKEN' ...
```

### 问题 3: CMSys 同步失败

**原因**: 配置文件中的接口地址错误

**解决**: 检查 `rpc/etc/cmpool.yaml` 中的配置
```yaml
CMSysDataSource:
  AuthEndpoint: "http://localhost:8888/platform/cmsys/auth"
  DataEndpoint: "http://localhost:8888/platform/cmsys/data"
```

## 日志说明

启动服务后，可以在日志中看到 Mock 接口的调用信息：

```
收到Mock CMSys认证请求
Mock CMSys认证 - AppCode: DB

收到Mock CMSys数据查询请求
Mock CMSys数据查询 - Query: , Token: mock-cmsys-token-..., Operator: admin
生成Mock CMSys响应 - 主机数量: 10

收到Mock CMSys基于IP列表数据查询请求
Mock CMSys基于IP列表查询 - Token: mock-cmsys-token-..., Operator: admin
Mock CMSys基于IP列表查询 - IP数量: 3
生成Mock CMSys基于IP列表响应 - 主机数量: 3
```

## 更新日志

### v1.1.0 (2025-01-21)
- 新增基于IP列表查询的Mock接口
- 支持通过POST请求指定IP列表
- 增强错误处理和参数验证

### v1.0.0 (2025-01-15)
- 首次实现 CMSys Mock 接口
- 支持认证接口模拟
- 支持数据接口模拟
- 自动生成 10 台主机的模拟数据
- 支持 remark 字段
