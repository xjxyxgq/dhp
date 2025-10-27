package model

// ClusterInfo 集群信息结构体
type ClusterInfo struct {
	ClusterName      string `db:"cluster_name"`
	ClusterGroupName string `db:"cluster_group_name"`
}