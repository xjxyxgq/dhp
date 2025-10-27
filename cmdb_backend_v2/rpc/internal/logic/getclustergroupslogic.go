package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetClusterGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterGroupsLogic {
	return &GetClusterGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}



// 查询所有集群组及对应业务线
func (l *GetClusterGroupsLogic) GetClusterGroups(in *cmpool.ClusterGroupsReq) (*cmpool.ClusterGroupsResp, error) {
	l.Logger.Info("查询集群组信息")

	// 从数据库查询集群组数据
	clusterGroups, err := l.getClusterGroupsFromDB()
	if err != nil {
		l.Logger.Errorf("查询集群组失败: %v", err)
		return &cmpool.ClusterGroupsResp{
			Success: false,
			Message: fmt.Sprintf("查询集群组失败: %v", err),
		}, nil
	}

	l.Logger.Infof("查询到%d个集群组", len(clusterGroups))
	return &cmpool.ClusterGroupsResp{
		Success:      true,
		Message:      "查询成功",
		ClusterGroup: clusterGroups,
	}, nil
}

// getClusterGroupsFromDB 从数据库查询集群组数据
func (l *GetClusterGroupsLogic) getClusterGroupsFromDB() ([]*cmpool.ClusterGroup, error) {
	// 查询所有有效的集群组
	rows, err := l.svcCtx.ClusterGroupsModel.FindAllClusterGroups(l.ctx)
	if err != nil {
		return nil, err
	}

	// 将数据库结果转换为protobuf结构
	var clusterGroups []*cmpool.ClusterGroup
	for _, row := range rows {
		// 处理 NULL 值，设置默认值为 "未知"
		groupName := "未知"
		if row.GroupName.Valid {
			groupName = row.GroupName.String
		}
		
		clusterName := "未知"
		if row.ClusterName.Valid {
			clusterName = row.ClusterName.String
		}
		
		departmentName := "未知"
		if row.DepartmentLineName.Valid {
			departmentName = row.DepartmentLineName.String
		}
		
		group := &cmpool.ClusterGroup{
			Id:             row.Id,
			CreateAt:       row.CreateTime,
			UpdateAt:       row.UpdateTime,
			GroupName:      groupName,
			ClusterName:    clusterName,
			DepartmentName: departmentName,
		}
		clusterGroups = append(clusterGroups, group)
	}

	return clusterGroups, nil
}
