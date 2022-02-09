// Package adapter adapter的配置工具
package adapter

import (
	"context"
	"errors"

	"github.com/GoAdminGroup/go-admin/engine"
	models2 "github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/kataras/iris/v12"
	"github.com/scjtqs2/bot_adapter/client"
	"github.com/scjtqs2/bot_adapter/config"
	"github.com/scjtqs2/bot_adapter/pb/entity"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/models"
	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/utils/common"
	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/utils/jump"
)

// Info todo
type Info struct {
	Cli         *client.AdapterService
	Conf        *models.AdapterConfig
	Status      bool // 是否启用
	AdapterConf *config.Config
}

// CheckAuth 校验是否登录
func CheckAuth(ctx iris.Context) error {
	var user models2.UserModel
	var ok bool
	user, ok = engine.User(ctx)
	if !ok {
		jump.ErrorForIris(ctx, common.Msg{
			Msg:  "获取登录信息失败",
			URL:  "/admin/",
			Wait: 3,
		})
		return errors.New("获取登录信息失败")
	}
	isAdmin := false
	for _, v := range user.Roles {
		if v.Id == 1 {
			isAdmin = true
		}
	}
	if !isAdmin {
		jump.ErrorForIris(ctx, common.Msg{
			Msg:  "非超级管理员",
			URL:  "/admin/",
			Wait: 3,
		})
		return errors.New("非超级管理员")
	}
	return nil
}

// NewInfo todo
func NewInfo() *Info {
	conf, err := models.GetAdapterConfig()
	if err != nil {
		return &Info{
			Status: false,
		}
	}
	return &Info{
		Conf:   conf,
		Status: true,
	}
}

// GetCli 获取/初始化adapter的grpc client
func (i *Info) GetCli() (*client.AdapterService, error) {
	if i.Cli != nil {
		return i.Cli, nil
	}
	var err error
	i.Cli, err = client.NewAdapterServiceClient(i.Conf.AdapterGrpcAdd, i.Conf.AppID, i.Conf.AppSecret)
	if err != nil {
		return nil, err
	}
	return i.Cli, nil
}

// getConfig 通过grpc拉取 bot-adapter的 yaml的配置信息
func (i *Info) getConfig() (*config.Config, error) {
	cli, err := i.GetCli()
	if err != nil {
		log.Errorf("init bot-adapter client err:%v", err)
		return nil, err
	}
	conf, err := cli.GetConfig(context.TODO(), &entity.Config{})
	if err != nil {
		log.Errorf("get bot-adapter config err:%v", err)
		return nil, err
	}
	err = yaml.Unmarshal(conf.GetConfig(), &i.AdapterConf)
	if err != nil {
		return nil, err
	}
	return i.AdapterConf, nil
}

// saveAdapterConfig 远程保存
func (i *Info) saveAdapterConfig(config *config.Config) (*entity.Config, error) {
	data, _ := yaml.Marshal(config)
	return i.Cli.UpdateConfig(context.TODO(), &entity.Config{Config: data})
}

// RebootAdatper 重启 bot-adapter
func (i *Info) RebootAdatper(ctx iris.Context) (types.Panel, error) {
	var err error
	if err = CheckAuth(ctx); err != nil {
		return types.Panel{}, nil
	}
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/info"
	} else if ctx.GetReferrer().Path == "/admin/adapter/applist" {
		re = "/admin/qq/info"
	}
	_, err = i.Cli.SetRestart(context.TODO(), &entity.SetRestartReq{Delay: 2})
	if err != nil {
		return jump.Error(common.Msg{
			Msg: err.Error(),
			URL: re,
		}), nil
	}
	return jump.Success(common.Msg{
		Msg: "操作成功",
		URL: re,
	}), nil
}
