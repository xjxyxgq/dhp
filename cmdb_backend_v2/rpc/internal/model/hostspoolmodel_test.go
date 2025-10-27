package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostsPoolModel_Basic(t *testing.T) {
	// 这是一个基本的模型创建测试
	// 由于我们没有实际的数据库连接，这里只测试模型的基本结构
	
	// 测试 HostsPool 结构体的基本字段
	host := &HostsPool{
		HostName: "test-host",
		HostIp:   "192.168.1.100",
	}
	
	assert.NotNil(t, host)
	assert.Equal(t, "test-host", host.HostName)
	assert.Equal(t, "192.168.1.100", host.HostIp)
}

func TestHostInfo_Basic(t *testing.T) {
	// 测试 HostInfo 结构体
	hostInfo := &HostInfo{
		HostIp:   "192.168.1.100",
		HostName: "test-host",
	}
	
	assert.NotNil(t, hostInfo)
	assert.Equal(t, "192.168.1.100", hostInfo.HostIp)
	assert.Equal(t, "test-host", hostInfo.HostName)
}

func TestHostPoolDetailRow_Basic(t *testing.T) {
	// 测试 HostPoolDetailRow 结构体
	detail := &HostPoolDetailRow{
		Id:       1,
		HostName: "test-host",
		HostIp:   "192.168.1.100",
		DiskSize: 500,
		Ram:      16,
		Vcpus:    4,
	}
	
	assert.NotNil(t, detail)
	assert.Equal(t, int64(1), detail.Id)
	assert.Equal(t, "test-host", detail.HostName)
	assert.Equal(t, "192.168.1.100", detail.HostIp)
	assert.Equal(t, int32(500), detail.DiskSize)
	assert.Equal(t, int32(16), detail.Ram)
	assert.Equal(t, int32(4), detail.Vcpus)
}