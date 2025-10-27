package logic

import (
	"context"
	"database/sql"
	"regexp"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateIdcConfLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateIdcConfLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateIdcConfLogic {
	return &CreateIdcConfLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建IDC机房配置
func (l *CreateIdcConfLogic) CreateIdcConf(in *cmpool.CreateIdcConfReq) (*cmpool.CreateIdcConfResp, error) {
	// 验证输入参数
	if in.IdcName == "" {
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "IDC机房名称不能为空",
		}, nil
	}
	
	if in.IdcCode == "" {
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "IDC机房代码不能为空",
		}, nil
	}
	
	if in.IdcIpRegexp == "" {
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "IP正则表达式不能为空",
		}, nil
	}

	// 验证正则表达式格式
	_, err := regexp.Compile(in.IdcIpRegexp)
	if err != nil {
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "无效的正则表达式格式",
		}, nil
	}

	// 检查IDC代码是否已存在
	existing, err := l.svcCtx.IdcConfModel.FindOneByIdcCode(l.ctx, in.IdcCode)
	if err != nil && err != sql.ErrNoRows {
		l.Logger.Errorf("查询IDC代码失败: %v", err)
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "系统错误，请稍后再试",
		}, nil
	}
	
	if existing != nil {
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "IDC机房代码已存在",
		}, nil
	}

	// 创建IDC配置
	idcConf := &model.IdcConf{
		IdcName:     in.IdcName,
		IdcCode:     in.IdcCode,
		IdcIpRegexp: in.IdcIpRegexp,
		IsActive:    1, // 默认激活
		Priority:    uint64(in.Priority),
	}
	
	if in.IdcLocation != "" {
		idcConf.IdcLocation = sql.NullString{String: in.IdcLocation, Valid: true}
	}
	
	if in.IdcDescription != "" {
		idcConf.IdcDescription = sql.NullString{String: in.IdcDescription, Valid: true}
	}

	result, err := l.svcCtx.IdcConfModel.Insert(l.ctx, idcConf)
	if err != nil {
		l.Logger.Errorf("创建IDC配置失败: %v", err)
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "创建IDC配置失败",
		}, nil
	}

	idcId, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("获取插入ID失败: %v", err)
		return &cmpool.CreateIdcConfResp{
			Success: false,
			Message: "创建IDC配置失败",
		}, nil
	}

	return &cmpool.CreateIdcConfResp{
		Success: true,
		Message: "创建IDC配置成功",
		IdcId:   idcId,
	}, nil
}
