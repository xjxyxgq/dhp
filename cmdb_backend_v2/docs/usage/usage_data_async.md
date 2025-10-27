目前项目中“监控指标数据加载”功能是通过上传服务器指标监控CSV文件的方式实现的，我还有一个在线的数据源是 ES，可以通过 phoenix 代理的 es 接口来查询数据并加载到 server_resource 表里去，这个接口的查询语法完全遵循 es 的查询语法，接口地址是 http://phoenix.local.com/platform/query/es ，接口应当是一个可配置的变量。

当前暂无这个接口的环境，请 mock 该接口的数据。已知 ES 中服务器用量指标存储在 cluster*:data-zabbix-host-monitor-* 这个 index 中，该 index 存储的数据中，分别包含以下字段：hostIp、hostName、cpu、available_memory、total_disk_space_all，对应到我们所需的 hostIP、hostName、MaxCpu、MaxMem、MaxDisk，你应当访问 ES 时，通过 hostIp 和 指定的查询时间范围来对数值数据进行聚合。其中 cpu、available_memory、total_disk_space_all 是聚合的数据，其中有以下 json 结构的 key： value_min value_max value_avg min max zabbix_value_count，以及这些 key 对应的具体浮点数值。

当实现了数据通过 ES 接口加载的功能后，还需要实现如下接口，为后续实现前端页面提供支持：
1. 需要支持将这个数据同步任务配置为一个定时任务，定期运行；
2. 记录定期运行的结果，包括运行成功与否、哪些 hosts_pool 中的服务器同步成功、哪些服务器同步失败、哪些 ES 的数据源中有主机，但 hosts_pool 中没有主机；
3. 支持通过接口展示定期运行结果的列表，并展示定期运行结果的详情；
4. 支持根据一个主机列表，并立即发起从 ES 抓取用量数据的任务，主机列表可能由前端通过文件附件的方式提供。

注意，当前项目的 cmdb_frontend_v2 已经废弃，当前实现不需要实现任何前端功能，请在实现接口后，提供具体的接口文档，供前端项目使用。