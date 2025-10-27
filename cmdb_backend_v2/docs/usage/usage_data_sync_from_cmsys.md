目前项目中“监控指标数据加载”功能已经可以从 csv 文件、从 es 加载。还需要实现另外一个数据来源的数据加载。这个数据源是一个 HTTP 接口，会返回一个 json 数据。

1. 接口地址依然是一个可配置的地址，它使用 https 协议以 GET 方式向目标地址发起请求，并通过在请求 url 中的参数 query 来配置具体接口的查询条件，这个接口的访问需要在 http header 中提供 x-control-access-token、x-control-access-operator 来做接口访问认证，其中 x-control-access-operator 的值通过参数配置，token 通过接口获得；
2. token 需要使用一个 https 协议以 POST 方式请求认证接口获得，接口所需的参数通过 POST 传递请求体 {"appCode": "DB", "secret": "xxx"} 获得，预期的认证结果为 {"code": "A0000", "msg": "success", "data": "xxxxxx"}，当请求成功时，data 即为所需的 token
3. 上述获取 token 过程中所需的 appCode、secret 值通过参数文件配置
4. 数据源接口的数据返回格式如下，实际接口返回值还有更多字段，以下是我们所需要的
{ "code": "A000", "msg": "success", "data": [
    {
        "ipAddress": "xxx",
        "cpuMaxNew": "11.11",
        "memMaxNew": "11.11",
        "diskMaxNew": "11.11",
        "remark": "xxxxxx"
    }
]}
其中 ipAddress 是主机 IP 地址，cpuMaxNew 是最大 CPU 利用率，memMaxNew 是最大内存利用率，diskMaxNew 是最大硬盘利用率，remark 是对这款机器进行解释说明的备注，例如它正在被下线中。获取到这些数据后，使用这些数据来更新 server_resource 中的数据。与现在的 ES 数据获取逻辑一样，需要将这个检索结果中的所有主机 IP 写入 hosts_pool
   表，将检索结果中的用量指标数据全部写入 server_resource 表，检索结果中存在不在当前 hosts_pool 中的主机，应当提示出来。
5. 请将抓取到的 remark 信息写入到 hosts_pool 的 remark 列中，我已经添加了这列，你需要更新对应的 model 代码，并在调用 /api/v1/hardware-proxy/cmdb/v1/get_hosts_pool_detail 接口时，返回主机的 remark 数据，便于后续使用

请在实现接口后，提供具体的接口文档，供前端项目使用。