# CMSys Mock 接口实现总结

## 实现概述

本次实现为 CMSys 数据同步功能添加了 Mock 接口支持，使得开发人员可以在不连接真实 CMSys 系统的情况下进行开发和测试。

## 实现内容

### 1. API 定义更新

**文件**: `api/cmdb.api`

在无需认证的服务部分添加了两个 Mock 接口：

```go
service cmdb-api {
    @handler MockCMSysAuth
    post /platform/cmsys/auth

    @handler MockCMSysData
    get /platform/cmsys/data
}
```

### 2. Mock 认证接口实现

**文件**: `api/internal/handler/mockcmsysauthhandler.go`

**功能特点**:
- 接受 POST 请求，解析 JSON 请求体（appCode, secret）
- 返回固定的 mock token，方便测试
- 响应格式符合 CMSys 认证接口规范

**核心代码**:
```go
func MockCMSysAuthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 解析请求体
        var reqBody struct {
            AppCode string `json:"appCode"`
            Secret  string `json:"secret"`
        }
        json.NewDecoder(r.Body).Decode(&reqBody)

        // 返回固定 token
        mockToken := "mock-cmsys-token-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
        response := map[string]interface{}{
            "code": "A0000",
            "msg":  "success",
            "data": mockToken,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }
}
```

### 3. Mock 数据接口实现

**文件**: `api/internal/handler/mockcmsysdatahandler.go`

**功能特点**:
- 接受 GET 请求，从 header 中读取 token 和 operator
- 验证 token 是否存在（未提供则返回错误）
- 自动生成 10 台虚拟主机的模拟数据
- 每次请求生成随机的资源利用率数值
- 包含真实场景的 remark 描述

**核心代码**:
```go
func MockCMSysDataHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 获取请求头
        token := r.Header.Get("x-control-access-token")
        operator := r.Header.Get("x-control-access-operator")

        // 验证 token
        if token == "" {
            response := map[string]interface{}{
                "code": "A0001",
                "msg":  "未提供认证token",
                "data": nil,
            }
            json.NewEncoder(w).Encode(response)
            return
        }

        // 生成 mock 数据
        mockData := generateMockCMSysData()
        response := map[string]interface{}{
            "code": "A000",
            "msg":  "success",
            "data": mockData,
        }

        json.NewEncoder(w).Encode(response)
    }
}
```

**Mock 数据生成**:
- IP 地址: 192.168.1.1 - 192.168.1.10
- CPU 利用率: 40-90% (随机)
- 内存利用率: 50-90% (随机)
- 磁盘利用率: 30-90% (随机)
- Remark: 10 种不同的服务器描述

### 4. 配置更新

**文件**: `rpc/etc/cmpool.yaml`

更新了 CMSys 数据源配置，默认指向 Mock 接口：

```yaml
CMSysDataSource:
  # 开发测试环境使用Mock接口
  AuthEndpoint: "http://localhost:8888/platform/cmsys/auth"
  DataEndpoint: "http://localhost:8888/platform/cmsys/data"
  # 生产环境使用真实接口（注释掉Mock配置，取消下面的注释）
  # AuthEndpoint: "https://api.cmsys.example.com/auth"
  # DataEndpoint: "https://api.cmsys.example.com/data"
  AppCode: "DB"
  AppSecret: "your-app-secret-here"
  Operator: "admin"
  TimeoutSeconds: 60
```

### 5. 测试脚本

**文件**: `test_cmsys_mock.sh`

提供了自动化测试脚本，包含三个测试场景：

1. **测试认证接口**: 获取 token
2. **测试数据接口**: 使用 token 查询数据
3. **测试 token 验证**: 验证无 token 请求被拒绝

**使用方式**:
```bash
cd cmdb_backend_v2
./test_cmsys_mock.sh
```

### 6. 文档

创建了两份文档：

**CMSYS_MOCK_INTERFACES.md**:
- Mock 接口详细使用说明
- 配置切换指南
- 完整测试流程
- Mock 数据说明
- 故障排查指南

**CMSYS_SYNC_API_DOCUMENTATION.md** (更新):
- 在原有文档中添加了 Mock 接口章节
- 说明如何使用 Mock 接口进行开发测试

## 技术实现特点

### 1. 与 ES Mock 一致的实现方式

参照现有的 `mockesqueryhandler.go` 实现方式：
- 在 handler 中直接实现 mock 逻辑
- 不使用 logic 层
- 使用 `map[string]interface{}` 构造响应

### 2. 符合 CMSys 接口规范

严格按照文档中的 CMSys 接口规范实现：
- 认证接口：POST 请求，返回 token
- 数据接口：GET 请求 + Header 认证
- 响应格式：包含 code, msg, data 字段

### 3. 真实场景模拟

Mock 数据尽可能接近真实场景：
- 合理的资源利用率范围
- 真实的服务器描述
- 完整的数据字段

### 4. 开发友好

- 固定 token，方便测试
- 自动生成数据，无需手动准备
- 详细的日志输出
- 提供自动化测试脚本

## 验证测试

### 编译验证

```bash
cd api
go build -o cmdb-api .
```

✅ 编译成功，无错误

### 代码检查

- ✅ 正确处理 JSON 请求和响应
- ✅ 正确验证 HTTP header
- ✅ 符合 go-zero 框架规范
- ✅ 代码注释完整

### 配置验证

- ✅ 配置文件格式正确
- ✅ Mock 接口地址已配置
- ✅ 支持快速切换到生产环境

## 使用流程

### 开发测试环境

1. **启动服务**:
   ```bash
   ./start.sh
   ```

2. **测试 Mock 接口**:
   ```bash
   ./test_cmsys_mock.sh
   ```

3. **测试完整同步**:
   ```bash
   curl -X POST 'http://localhost:8888/api/cmdb/v1/cmsys-sync' \
     -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer YOUR_TOKEN' \
     -d '{"task_name": "Mock测试"}'
   ```

### 生产环境

修改 `rpc/etc/cmpool.yaml`，切换到真实接口地址即可。

## 与 ES Mock 的对比

| 特性 | CMSys Mock | ES Mock |
|------|-----------|---------|
| 认证 | ✅ 需要 token | ❌ 无需认证 |
| 接口数量 | 2 个（auth + data） | 1 个 |
| 请求方式 | POST + GET | POST |
| 数据格式 | JSON array | ES aggregation |
| Token 验证 | ✅ 有 | ❌ 无 |
| 主机数量 | 固定 10 台 | 动态 |

## 文件清单

### 新增文件

```
api/internal/handler/
├── mockcmsysauthhandler.go    # 认证接口 handler
└── mockcmsysdatahandler.go    # 数据接口 handler

api/internal/logic/
├── mockcmsysauthlogic.go      # 认证接口 logic（未使用）
└── mockcmsysdatalogic.go      # 数据接口 logic（未使用）

cmdb_backend_v2/
├── test_cmsys_mock.sh                      # 测试脚本
├── CMSYS_MOCK_INTERFACES.md                # Mock 接口使用说明
└── CMSYS_MOCK_IMPLEMENTATION_SUMMARY.md    # 实现总结（本文件）
```

### 修改文件

```
api/cmdb.api                    # 添加 Mock 接口路由
rpc/etc/cmpool.yaml             # 更新配置指向 Mock 接口
CMSYS_SYNC_API_DOCUMENTATION.md # 添加 Mock 接口说明
```

## 后续建议

1. **功能增强**:
   - 支持通过 query 参数过滤返回的主机
   - 支持配置主机数量和 IP 段
   - 支持自定义资源利用率范围

2. **测试覆盖**:
   - 添加单元测试
   - 添加集成测试
   - 性能测试

3. **监控告警**:
   - 添加 Mock 接口调用统计
   - 生产环境禁用 Mock 接口的检测机制

4. **文档完善**:
   - 添加视频教程
   - 添加常见问题 FAQ
   - 添加开发最佳实践

## 总结

本次实现为 CMSys 数据同步功能提供了完整的 Mock 接口支持，包括：

✅ 2 个 Mock HTTP 接口（认证 + 数据）
✅ 符合 CMSys 接口规范
✅ 完整的测试脚本
✅ 详细的使用文档
✅ 开发测试配置
✅ 编译通过，无错误

开发人员现在可以在不依赖真实 CMSys 系统的情况下，快速进行 CMSys 数据同步功能的开发和测试。
