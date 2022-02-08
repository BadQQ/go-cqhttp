package qq

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	db2 "github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/paginator"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	tmpl "github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/icon"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"github.com/GoAdminGroup/themes/adminlte/components/infobox"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/kataras/iris/v12"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/loghook"
	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/utils/common"
	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/utils/jump"
	"github.com/Mrs4s/go-cqhttp/db"
	"github.com/Mrs4s/go-cqhttp/global"
)

// QqInfo 查看当前的qq状态
func (l *Dologin) QqInfo(ctx iris.Context) (types.Panel, error) {
	components := tmpl.Get(config.GetTheme())
	colComp := components.Col()
	/**************************
	 * Info Box
	/**************************/

	infobox1 := infobox.New().
		SetText("QQ账号").
		SetColor("aqua").
		SetNumber(tmpl.HTML(strconv.FormatInt(l.Cli.Uin, 10))).
		SetIcon("ion-ios-gear-outline").
		GetContent()
	status := func(l *Dologin) string {
		if l.Cli.Online.Load() {
			return "QQ在线"
		}
		return "QQ离线"
	}(l)
	infobox2 := infobox.New().
		SetText("QQ状态").
		SetColor("red").
		SetNumber(tmpl.HTML(status)).
		SetIcon(icon.GooglePlus).
		SetContent(tmpl.HTML(func() string {
			if l.Cli.Online.Load() {
				return fmt.Sprintf("登录协议：%s", client.SystemDeviceInfo.Protocol)
			}
			return ""
		}())).
		GetContent()
	serverStatus := func(l *Dologin) string {
		if l.Status {
			return "server已启动"
		}
		return "server服务未启动"
	}(l)
	infobox3 := infobox.New().
		SetText("Server状态").
		SetColor("green").
		SetNumber(tmpl.HTML(serverStatus)).
		SetIcon("ion-ios-cart-outline").
		GetContent()

	infobox4 := infobox.New().
		SetText("错误信息").
		SetColor("yellow").
		SetNumber(tmpl.HTML(fmt.Sprintf("code:%d ,msg:%s", l.ErrMsg.Code, l.ErrMsg.Msg))).
		SetIcon("ion-ios-people-outline"). // svg is ok
		GetContent()

	var size = types.SizeMD(3).SM(6).XS(12)
	infoboxCol1 := colComp.SetSize(size).SetContent(infobox1).GetContent()
	infoboxCol2 := colComp.SetSize(size).SetContent(infobox2).GetContent()
	infoboxCol3 := colComp.SetSize(size).SetContent(infobox3).GetContent()
	infoboxCol4 := colComp.SetSize(size).SetContent(infobox4).GetContent()
	row1 := components.Row().SetContent(infoboxCol1 + infoboxCol2 + infoboxCol3 + infoboxCol4).GetContent()

	lab1 := components.Label().SetContent("快捷操作：").SetType("warning").GetContent()
	rowlab := components.Row().SetContent(tmpl.Default().Box().WithHeadBorder().SetBody(lab1).GetContent()).GetContent()
	link1 := components.Link().
		SetURL("/admin/info/qq_config"). // 设置跳转路由
		SetContent("修改配置信息"). // 设置链接内容
		SetTabTitle("修改配置信息").
		SetClass("btn btn-sm btn-danger").
		GetContent()

	link2 := components.Link().
		SetURL("/admin/qq/checkconfig").
		SetContent("开始登录").
		SetTabTitle("开始登录").
		SetClass("btn btn-sm btn-primary").
		GetContent()

	link3 := components.Link().
		SetURL("/admin/qq/shutdown").
		SetContent("关闭服务").
		SetTabTitle("关闭服务").
		SetClass("btn btn-sm btn-primary").
		GetContent()
	linkinfo := components.Link().
		SetURL("/admin/application/info").
		SetContent("系统信息").
		SetTabTitle("系统信息").
		SetClass("btn btn-sm btn-default").
		GetContent()
	linkLog := components.Link().
		SetURL("/admin/qq/weblog").
		OpenInNewTab().
		SetContent("系统日志").
		SetTabTitle("系统日志").
		SetClass("btn btn-sm btn-default").
		GetContent()
	linkChangeUserinfo := components.Link().
		SetURL("/admin/info/normal_manager/edit?__goadmin_edit_pk=1").
		SetContent("修改后台登录密码").
		SetClass("btn btn-sm btn-default").
		GetContent()
	linkHelp := components.Link().
		SetURL("/admin/qq/help").
		SetContent("使用帮助").
		SetClass("btn btn-sm btn-default").
		GetContent()
	linkDeviceInfo := components.Link().
		SetURL("/admin/qq/deviceinfo").
		SetContent("修改device.json").
		SetClass("btn btn-sm btn-default").
		GetContent()
	rown1 := components.Row().SetContent(tmpl.Default().Box().WithHeadBorder().
		SetBody(
			link1 + link2 + link3 + linkinfo + linkLog + linkChangeUserinfo + linkHelp + linkDeviceInfo,
		).GetContent()).GetContent()
	link4 := components.Link().
		SetURL("/admin/qq/friendlist").
		SetContent("好友列表").
		SetClass("btn btn-sm btn-default").
		GetContent()
	link5 := components.Link().
		SetURL("/admin/qq/grouplist").
		SetContent("群组列表").
		SetClass("btn btn-sm btn-default").
		GetContent()
	linkGuildList := components.Link().
		SetURL("/admin/qq/guildlist").
		SetContent("频道列表").
		SetClass("btn btn-sm btn-default").
		GetContent()
	rown2 := components.Row().SetContent(tmpl.Default().Box().WithHeadBorder().SetBody(link4 + link5 + linkGuildList).GetContent()).GetContent()

	return types.Panel{
		Content:     row1 + rowlab + rown1 + rown2,
		Title:       "qq状态",
		Description: "当前qq状态信息",
	}, nil
}

// MemberList 好友列表
func (l *Dologin) MemberList(ctx iris.Context) (types.Panel, error) {
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	comp := tmpl.Get(config.GetTheme())
	fs := make([]map[string]types.InfoItem, 0, len(l.Cli.FriendList))
	for _, f := range l.Cli.FriendList {
		linkDelete := comp.Link().SetClass("btn btn-sm btn-danger").SetContent("删除").SetURL("/admin/qq/deletefriend?uin=" + strconv.FormatInt(f.Uin, 10)).GetContent()
		linkDetail := comp.Link().SetClass("btn btn-sm btn-primary").SetContent("详情").SetURL("/admin/qq/getfrienddetail?uin=" + strconv.FormatInt(f.Uin, 10)).GetContent()
		linkSendMsg := comp.Link().SetClass("btn btn-sm btn-default").SetContent("和TA聊天").SetURL("/admin/qq/getmsglist?uin=" + strconv.FormatInt(f.Uin, 10)).GetContent()
		fs = append(fs, map[string]types.InfoItem{
			"昵称":  {Content: tmpl.HTML(f.Nickname), Value: f.Nickname},
			"备注":  {Content: tmpl.HTML(f.Remark), Value: f.Remark},
			"QQ号": {Content: tmpl.HTML(strconv.FormatInt(f.Uin, 10)), Value: strconv.FormatInt(f.Uin, 10)},
			"操作":  {Content: linkSendMsg + linkDetail + linkDelete},
		})
	}

	param := parameter.GetParam(ctx.Request().URL, 300)
	table := comp.Table().
		SetInfoList(common.PageSlice(fs, param.PageInt, param.PageSizeInt)).
		SetThead(types.Thead{
			{Head: "昵称", Width: "15%"},
			{Head: "备注", Width: "15%"},
			{Head: "QQ号", Width: "10%"},
			{Head: "操作"},
		}).SetMinWidth("0.01%")

	body := table.GetContent()

	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetNoPadding().
			WithHeadBorder().
			SetFooter(paginator.Get(paginator.Config{
				Size:         len(l.Cli.FriendList),
				PageSizeList: []string{"100", "200", "300", "500"},
				Param:        param,
			}).GetContent()).
			GetContent(),
		Title:       "好友列表",
		Description: tmpl.HTML(fmt.Sprintf("%d的好友列表", l.Cli.Uin)),
	}, nil
}

// GroupList 群组列表
func (l *Dologin) GroupList(ctx iris.Context) (types.Panel, error) {
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	fs := make([]map[string]types.InfoItem, 0, len(l.Cli.GroupList))
	comp := tmpl.Get(config.GetTheme())

	for _, g := range l.Cli.GroupList {
		linkLeave := comp.Link().SetClass("btn btn-sm btn-danger").SetContent("退出群").SetURL("/admin/qq/leavegroup?guin=" + strconv.FormatInt(g.Code, 10)).GetContent()
		linkDetail := comp.Link().SetClass("btn btn-sm btn-primary").SetContent("详情").SetURL("/admin/qq/getgroupdetail?guin=" + strconv.FormatInt(g.Code, 10)).GetContent()
		linkSendMsg := comp.Link().SetClass("btn btn-sm btn-default").SetContent("进去聊天").SetURL("/admin/qq/getgroupmsglist?uin=" + strconv.FormatInt(g.Code, 10)).GetContent()
		fs = append(fs, map[string]types.InfoItem{
			"群组名": {Content: tmpl.HTML(g.Name)},
			"群号":  {Content: tmpl.HTML(strconv.FormatInt(g.Code, 10))},
			"操作":  {Content: linkSendMsg + linkDetail + linkLeave},
		})
	}
	param := parameter.GetParam(ctx.Request().URL, 100)

	table := comp.Table().
		SetInfoList(common.PageSlice(fs, param.PageInt, param.PageSizeInt)).
		SetThead(types.Thead{
			{Head: "群组名", Width: "30%"},
			{Head: "群号", Width: "15%"},
			{Head: "操作"},
		}).SetMinWidth("0.01%")

	body := table.GetContent()

	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetNoPadding().
			WithHeadBorder().
			SetFooter(paginator.Get(paginator.Config{
				Size:         len(l.Cli.GroupList),
				PageSizeList: []string{"100", "200", "300", "500"},
				Param:        param,
			}).GetContent()).
			GetContent(),
		Title:       "群组列表",
		Description: tmpl.HTML(fmt.Sprintf("%d的群组列表", l.Cli.Uin)),
	}, nil
}

// GuildList 频道guild列表
func (l *Dologin) GuildList(ctx iris.Context) (types.Panel, error) {
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	comp := tmpl.Get(config.GetTheme())
	fs := make([]map[string]types.InfoItem, 0, len(l.Cli.GuildService.Guilds))
	for _, info := range l.Cli.GuildService.Guilds {
		// linkLeave := comp.Link().SetClass("btn btn-sm btn-danger").SetContent("退出频道").SetURL("/admin/qq/leaveguild?guid=" + strconv.FormatUint(info.GuildId, 10)).GetContent()
		linkList := comp.Link().SetClass("btn btn-sm btn-primary").SetContent("查看子频道").SetURL("/admin/qq/channellist?guid=" + strconv.FormatUint(info.GuildId, 10)).GetContent()
		// linkSendMsg := comp.Link().SetClass("btn btn-sm btn-default").SetContent("进去聊天").SetURL("/admin/qq/getgroupmsglist?uin=" + strconv.FormatInt(g.Code, 10)).GetContent()
		fs = append(fs, map[string]types.InfoItem{
			"频道名": {Content: tmpl.HTML(info.GuildName)},
			"频道号": {Content: tmpl.HTML(strconv.FormatUint(info.GuildId, 10))},
			"操作":  {Content: linkList},
		})
	}
	param := parameter.GetParam(ctx.Request().URL, 100)
	table := comp.Table().
		SetInfoList(common.PageSlice(fs, param.PageInt, param.PageSizeInt)).
		SetThead(types.Thead{
			{Head: "频道名", Width: "30%"},
			{Head: "频道号", Width: "15%"},
			{Head: "操作"},
		}).SetMinWidth("0.01%")
	body := table.GetContent()
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetNoPadding().
			WithHeadBorder().
			SetFooter(paginator.Get(paginator.Config{
				Size:         len(l.Cli.GuildService.Guilds),
				PageSizeList: []string{"100", "200", "300", "500"},
				Param:        param,
			}).GetContent()).
			GetContent(),
		Title:       "频道列表",
		Description: tmpl.HTML(fmt.Sprintf("%d的频道列表", l.Cli.Uin)),
	}, nil
}

// ChannelList 子频道列表
func (l *Dologin) ChannelList(ctx iris.Context) (types.Panel, error) {
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	comp := tmpl.Get(config.GetTheme())
	fs := make([]map[string]types.InfoItem, 0, len(l.Cli.GuildService.Guilds))
	guildID, _ := strconv.ParseUint(ctx.URLParam("guid"), 10, 64)
	guild := l.Cli.GuildService.FindGuild(guildID)
	if guild == nil {
		return jump.Error(common.Msg{
			Msg: "GUILD_NOT_FOUND",
			URL: "/admin/qq/guildlist",
		}), nil
	}
	guild.Channels, err = l.Cli.GuildService.FetchChannelList(guildID)
	if err != nil {
		log.Errorf("获取频道 %v 子频道列表时出现错误: %v", guildID, err)
		return jump.Error(common.Msg{
			Msg: "API_ERROR" + err.Error(),
			URL: "/admin/qq/guildlist",
		}), nil
	}
	for _, c := range guild.Channels {
		// linkLeave := comp.Link().SetClass("btn btn-sm btn-danger").SetContent("退出频道").SetURL("/admin/qq/leavegroup?guin=" + strconv.FormatUint(c.ChannelId, 10)).GetContent()
		linkSendMsg := comp.Link().SetClass("btn btn-sm btn-primary").SetContent("进入子频道").SetURL(fmt.Sprintf("/admin/qq/getguildchennelmsglist?guid=%d&cid=%d", guildID, c.ChannelId)).GetContent()
		// linkSendMsg := comp.Link().SetClass("btn btn-sm btn-default").SetContent("进去聊天").SetURL("/admin/qq/getgroupmsglist?uin=" + strconv.FormatInt(g.Code, 10)).GetContent()
		fs = append(fs, map[string]types.InfoItem{
			"子频道名": {Content: tmpl.HTML(c.ChannelName)},
			"子频道号": {Content: tmpl.HTML(strconv.FormatUint(c.ChannelId, 10))},
			"操作":   {Content: linkSendMsg},
		})
	}
	param := parameter.GetParam(ctx.Request().URL, 100)
	table := comp.Table().
		SetInfoList(common.PageSlice(fs, param.PageInt, param.PageSizeInt)).
		SetThead(types.Thead{
			{Head: "子频道名", Width: "30%"},
			{Head: "子频道号", Width: "15%"},
			{Head: "操作"},
		}).SetMinWidth("0.01%")
	body := table.GetContent()
	meta, err := l.Cli.GuildService.FetchGuestGuild(guildID)
	var guidName string
	if err != nil {
		guidName = ctx.URLParam("guid")
	} else {
		guidName = meta.GuildName
	}
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetNoPadding().
			WithHeadBorder().
			SetFooter(paginator.Get(paginator.Config{
				Size:         len(l.Cli.GuildService.Guilds),
				PageSizeList: []string{"100", "200", "300", "500"},
				Param:        param,
			}).GetContent()).
			GetContent(),
		Title:       tmpl.HTML(fmt.Sprintf("%s的子频道列表", guidName)),
		Description: tmpl.HTML(fmt.Sprintf("%d的子频道列表", guildID)),
	}, nil
}

// GetFriendDetal 获取好友详细资料
func (l *Dologin) GetFriendDetal(ctx iris.Context) (types.Panel, error) {
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/friendlist"
	} else if ctx.GetReferrer().Path == "/admin/qq/getfrienddetail" {
		re = "/admin/qq/info"
	}
	uin, err := ctx.URLParamInt64("uin")
	if err != nil {
		return jump.Error(common.Msg{
			Msg: "参数错误",
			URL: re,
		}), nil
	}
	info, err := l.Cli.GetSummaryInfo(uin)
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取uin%d的信息失败：%s", uin, err.Error()),
			URL: re,
		}), nil
	}
	comp := tmpl.Get(config.GetTheme())

	table := comp.Table().
		SetHideThead().
		SetStyle("striped").
		SetMinWidth("30").
		SetInfoList([]map[string]types.InfoItem{
			{
				"key": {Content: "uin"},
				"val": {Content: tmpl.HTML(strconv.FormatInt(info.Uin, 10))},
			},
			{
				"key": {Content: "nickname"},
				"val": {Content: tmpl.HTML(info.Nickname)},
			},
			{
				"key": {Content: "city"},
				"val": {Content: tmpl.HTML(info.City)},
			},
			{
				"key": {Content: "mobile"},
				"val": {Content: tmpl.HTML(info.Mobile)},
			},
			{
				"key": {Content: "qid"},
				"val": {Content: tmpl.HTML(info.Qid)},
			},
			{
				"key": {Content: "sign"},
				"val": {Content: tmpl.HTML(info.Sign)},
			},
			{
				"key": {Content: "age"},
				"val": {Content: tmpl.HTML(fmt.Sprintf("%d", info.Age))},
			},
			{
				"key": {Content: "level"},
				"val": {Content: tmpl.HTML(fmt.Sprintf("%d", info.Level))},
			},
			{
				"key": {Content: "login_days"},
				"val": {Content: tmpl.HTML(fmt.Sprintf("%d", info.LoginDays))},
			},
			{
				"key": {Content: "sex"},
				"val": {Content: tmpl.HTML(func() string {
					if info.Sex == 1 {
						return "female"
					} else if info.Sex == 0 {
						return "male"
					}
					// unknown = 0x2
					return "unknown"
				}())},
			},
		}).
		SetThead(types.Thead{
			{Head: "key"},
			{Head: "val"},
		})

	body := table.GetContent()
	linkBack := comp.Link().SetURL(re).SetContent("返回").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetFooter(linkBack).
			SetNoPadding().
			WithHeadBorder().
			GetContent(),
		Title:       "好友详细信息",
		Description: tmpl.HTML(fmt.Sprintf("%d的详细资料", l.Cli.Uin)),
	}, nil
}

// GetGroupDetal 获取群信息
func (l *Dologin) GetGroupDetal(ctx iris.Context) (types.Panel, error) {
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/grouplist"
	} else if ctx.GetReferrer().Path == "/admin/qq/getgroupdetail" {
		re = "/admin/qq/info"
	}
	guin, err := ctx.URLParamInt64("guin")
	if err != nil {
		return jump.Error(common.Msg{
			Msg: "参数错误",
			URL: re,
		}), nil
	}
	group := l.Cli.FindGroup(guin)
	if group == nil {
		group, _ = l.Cli.GetGroupInfo(guin)
	}
	if group == nil {
		gid := strconv.FormatInt(guin, 10)
		info, err := l.Cli.SearchGroupByKeyword(gid)
		if err != nil {
			return jump.Error(common.Msg{
				Msg: fmt.Sprintf("获取uin%d的信息失败：%s", guin, err.Error()),
				URL: re,
			}), nil
		}
		for _, g := range info {
			if g.Code == guin {
				group = &client.GroupInfo{
					Code:            g.Code,
					Name:            g.Name,
					Memo:            g.Memo,
					Uin:             0,
					OwnerUin:        0,
					GroupCreateTime: 0,
					GroupLevel:      0,
					MaxMemberCount:  0,
					MemberCount:     0,
					Members:         nil,
				}
			}
		}
	}
	if group == nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取uin%d的信息失败", guin),
			URL: re,
		}), nil
	}
	comp := tmpl.Get(config.GetTheme())

	table := comp.Table().
		SetHideThead().
		SetStyle("striped").
		SetMinWidth("30").
		SetInfoList([]map[string]types.InfoItem{
			{
				"key": {Content: "uin"},
				"val": {Content: tmpl.HTML(strconv.FormatInt(group.Uin, 10))},
			},
			{
				"key": {Content: "name"},
				"val": {Content: tmpl.HTML(group.Name)},
			},
			{
				"key": {Content: "memo"},
				"val": {Content: tmpl.HTML(group.Memo)},
			},
			{
				"key": {Content: "code"},
				"val": {Content: tmpl.HTML(strconv.FormatInt(group.Code, 10))},
			},
			{
				"key": {Content: "ownerUin"},
				"val": {Content: tmpl.HTML(strconv.FormatInt(group.OwnerUin, 10))},
			},
			{
				"key": {Content: "GroupCreateTime"},
				"val": {Content: tmpl.HTML(func() string {
					if group.GroupCreateTime == 0 {
						return ""
					}
					timeLayout := "2006-01-02 15:04:05"
					return time.Unix(int64(group.GroupCreateTime), 0).Format(timeLayout)
				}())},
			},
			{
				"key": {Content: "GroupLevel"},
				"val": {Content: tmpl.HTML(fmt.Sprintf("%d", group.GroupLevel))},
			},
			{
				"key": {Content: "MemberCount"},
				"val": {Content: tmpl.HTML(fmt.Sprintf("%d", group.MemberCount))},
			},
			{
				"key": {Content: "MaxMemberCount"},
				"val": {Content: tmpl.HTML(fmt.Sprintf("%d", group.MaxMemberCount))},
			},
		}).
		SetThead(types.Thead{
			{Head: "key"},
			{Head: "val"},
		})

	body := table.GetContent()
	linkBack := comp.Link().SetURL(re).SetContent("返回").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetFooter(linkBack).
			SetNoPadding().
			WithHeadBorder().
			GetContent(),
		Title:       "好友详细信息",
		Description: tmpl.HTML(fmt.Sprintf("%d的详细资料", l.Cli.Uin)),
	}, nil
}

// WebLog 最近的100条日志
func (l *Dologin) WebLog(ctx *context.Context) (types.Panel, error) {
	comp := tmpl.Get(config.GetTheme())
	re := ctx.Referer()
	if re == "" {
		re = "/admin/qq/info"
	} else if ctx.RefererURL().Path == "/admin/qq/weblog" {
		re = "/admin/qq/info"
	}
	refresh := ctx.QueryDefault("refresh", "true")
	body := comp.Box().SetBody(tmpl.HTML(l.Weblog.Read())).GetContent()
	linkBack := comp.Link().SetURL(re).SetContent("返回").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	linkStop := comp.Link().SetURL("?refresh=false").SetContent("停止刷新").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	linkSart := comp.Link().SetURL("?refresh=true").SetContent("启动刷新").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetFooter(linkBack + linkStop + linkSart).
			SetNoPadding().
			WithHeadBorder().
			GetContent(),
		Title:       "系统日志",
		Description: "go-cqhttp的最近100条日志。每2秒刷新",
		AutoRefresh: func() bool {
			return refresh == "true"
		}(),
		RefreshInterval: []int{2},
	}, nil
}

// GetMsgListAjaxHTML 通过ajax拉取消息记录
func (l *Dologin) GetMsgListAjaxHTML(ctx iris.Context) {
	type data struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		HTML template.HTML `json:"html"`
	}
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return
	}
	uin := ctx.URLParamInt64Default("uin", 0)
	guid, _ := strconv.ParseUint(ctx.URLParam("guid"), 10, 64)
	cid, _ := strconv.ParseUint(ctx.URLParam("cid"), 10, 64)
	if uin == 0 && (guid == 0 || cid == 0) {
		_, _ = ctx.JSON(data{Code: -1, Msg: "empty uin or guildid or channelid"})
	}
	var body template.HTML
	if uin != 0 {
		list := loghook.ReadMsg(uin)
		for _, v := range list {
			body += l.parseMsg(v)
		}
	} else {
		// 频道消息
		list := loghook.ReadGuildChannelMsg(guid, cid)
		for _, v := range list {
			body += l.parseGuildChannelMsg(v)
		}
	}
	_, _ = ctx.JSON(data{Code: 200, HTML: body})
}

// GetMsgList 拉取和好友的聊天页面
func (l *Dologin) GetMsgList(ctx iris.Context) (types.Panel, error) {
	components := tmpl.Get(config.GetTheme())
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	uin := ctx.URLParamInt64Default("uin", 0)
	if uin == 0 {
		return jump.Error(common.Msg{
			Msg: "参数错误",
			URL: "/admin/qq/info",
		}), nil
	}
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/friendlist"
	} else if ctx.GetReferrer().Path == "/admin/qq/getmsglist" {
		re = "/admin/qq/friendlist"
	}
	userinfo, err := l.Cli.GetSummaryInfo(uin)
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取uin%d的信息失败：%s", uin, err.Error()),
			URL: re,
		}), nil
	}
	list := loghook.ReadMsg(uin)
	var body template.HTML
	for _, v := range list {
		body += l.parseMsg(v)
	}
	box := components.Box().WithHeadBorder().SetHeader(tmpl.HTML(fmt.Sprintf("和 %s(%d) 的聊天记录", userinfo.Nickname, uin))).
		SetBody(body).
		SetFooter("").
		SetAttr(`id="msgbox"`).
		GetContent()
	linkBack := components.Link().SetContent("返回群列表").SetClass("btn btn-sm btn-default").SetURL("/admin/qq/friendlist").GetContent()
	buttonSend := components.Button().SetContent("发送").SetThemeDefault().SetID("msgsend").GetContent()
	texeArea := components.Box().SetBody(tmpl.HTML(fmt.Sprintf(`
<textarea class="form-control" id="msgtext" data-username="%s" data-type="private" rows="3"></textarea>
`, l.Cli.Nickname))).
		SetFooter(linkBack + buttonSend).
		GetContent()
	fs := common.GetStaticFs()
	js, _ := fs.ReadFile("js/getmsglist.js")
	return types.Panel{
		Content:     box + texeArea,
		Title:       tmpl.HTML(fmt.Sprintf("和%s(%d)的聊天", userinfo.Nickname, uin)),
		Description: tmpl.HTML(fmt.Sprintf("和%s(%d)的聊天", userinfo.Nickname, uin)),
		JS:          tmpl.JS(string(js)),
	}, nil
}

// GetGroupMsgList 群聊天记录列表
func (l *Dologin) GetGroupMsgList(ctx iris.Context) (types.Panel, error) {
	components := tmpl.Get(config.GetTheme())
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	uin := ctx.URLParamInt64Default("uin", 0)
	if uin == 0 {
		return jump.Error(common.Msg{
			Msg: "参数错误",
			URL: "/admin/qq/info",
		}), nil
	}
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/grouplist"
	} else if ctx.GetReferrer().Path == "/admin/qq/getgroupmsglist" {
		re = "/admin/qq/grouplist"
	}
	groupinfo, err := l.getGroupInfo(uin)
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取uin%d的信息失败：%s", uin, err.Error()),
			URL: re,
		}), nil
	}
	list := loghook.ReadMsg(uin)
	var body template.HTML
	for _, v := range list {
		body += l.parseMsg(v)
	}
	box := components.Box().WithHeadBorder().SetHeader(tmpl.HTML(fmt.Sprintf("和 %s 的聊天记录", groupinfo.Name))).
		SetBody(body).
		SetFooter("").
		SetAttr(`id="msgbox"`).
		GetContent()
	linkBack := components.Link().SetContent("返回群列表").SetClass("btn btn-sm btn-default").SetURL("/admin/qq/grouplist").GetContent()
	buttonSend := components.Button().SetContent("发送").SetThemeDefault().SetID("msgsend").GetContent()
	texeArea := components.Box().SetBody(tmpl.HTML(fmt.Sprintf(`
<textarea class="form-control" id="msgtext" data-username="%s" data-type="group" rows="3"></textarea>
`, l.Cli.Nickname))).
		SetFooter(linkBack + buttonSend).
		GetContent()
	fs := common.GetStaticFs()
	js, _ := fs.ReadFile("js/getmsglist.js")
	return types.Panel{
		Content:     box + texeArea,
		Title:       tmpl.HTML(fmt.Sprintf("和%s(%d)的聊天", groupinfo.Name, uin)),
		Description: tmpl.HTML(fmt.Sprintf("和%s(%d)的聊天", groupinfo.Name, uin)),
		JS:          tmpl.JS(string(js)),
	}, nil
}

// GetGuildChannelMsgList 频道聊天记录列表
func (l *Dologin) GetGuildChannelMsgList(ctx iris.Context) (types.Panel, error) {
	components := tmpl.Get(config.GetTheme())
	err := l.CheckQQlogin(ctx)
	if err != nil {
		return types.Panel{}, nil
	}
	guid, _ := strconv.ParseUint(ctx.URLParam("guid"), 10, 64)
	cid, _ := strconv.ParseUint(ctx.URLParam("cid"), 10, 64)
	if guid == 0 || cid == 0 {
		return jump.Error(common.Msg{
			Msg: "参数错误",
			URL: "/admin/qq/info",
		}), nil
	}
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/guildlist"
	} else if ctx.GetReferrer().Path == "/admin/qq/getguildchennelmsglist" {
		re = "/admin/qq/guildlist"
	}
	chanlelInfo, err := l.Cli.GuildService.FetchChannelInfo(guid, cid)
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取guid:%d channelid:%d的信息失败：%s", guid, cid, err.Error()),
			URL: re,
		}), nil
	}
	guildinfo := l.Cli.GuildService.FindGuild(guid)
	var guildname string
	if guildinfo == nil {
		guildname = strconv.FormatUint(guid, 10)
	} else {
		guildname = guildinfo.GuildName
	}
	list := loghook.ReadGuildChannelMsg(guid, cid)
	var body template.HTML
	for _, v := range list {
		body += l.parseGuildChannelMsg(v)
	}
	box := components.Box().WithHeadBorder().SetHeader(tmpl.HTML(fmt.Sprintf("和 %s 的聊天记录", chanlelInfo.ChannelName))).
		SetBody(body).
		SetFooter("").
		SetAttr(`id="msgbox"`).
		GetContent()
	linkBack := components.Link().SetContent("返回子频道列表").SetClass("btn btn-sm btn-default").SetURL(fmt.Sprintf("/admin/qq/channellist?guid=%d", guid)).GetContent()
	linkBack2 := components.Link().SetContent("返回频道列表").SetClass("btn btn-sm btn-default").SetURL("/admin/qq/guildlist").GetContent()
	buttonSend := components.Button().SetContent("发送").SetThemeDefault().SetID("msgsend").GetContent()
	texeArea := components.Box().SetBody(tmpl.HTML(fmt.Sprintf(`
<textarea class="form-control" id="msgtext" data-username="%s" data-type="channel" rows="3"></textarea>
`, l.Cli.Nickname))).
		SetFooter(linkBack2 + linkBack + buttonSend).
		GetContent()
	fs := common.GetStaticFs()
	js, _ := fs.ReadFile("js/getmsglist.js")
	return types.Panel{
		Content:     box + texeArea,
		Title:       tmpl.HTML(fmt.Sprintf("和%s(%d)/%s(%d)的聊天", guildname, guid, chanlelInfo.ChannelName, cid)),
		Description: tmpl.HTML(fmt.Sprintf("和%s(%d)/%s(%d)的聊天", guildname, guid, chanlelInfo.ChannelName, cid)),
		JS:          tmpl.JS(string(js)),
	}, nil
}

func (l *Dologin) getGroupInfo(groupID int64) (*client.GroupInfo, error) {
	group := l.Cli.FindGroup(groupID)
	if group == nil {
		group, _ = l.Cli.GetGroupInfo(groupID)
	}
	if group == nil {
		gid := strconv.FormatInt(groupID, 10)
		info, err := l.Cli.SearchGroupByKeyword(gid)
		if err != nil {
			return nil, err
		}
		for _, g := range info {
			if g.Code == groupID {
				return &client.GroupInfo{
					Code: g.Code,
					Name: g.Name,
					Memo: g.Memo,
				}, nil
			}
		}
	} else {
		return group, nil
	}
	return nil, errors.New("GROUP_NOT_FOUND")
}
func (l *Dologin) parseMsg(msg db.StoredMessage) template.HTML {
	var text template.HTML
	for _, v := range msg.GetContent() {
		data := v["data"].(global.MSG)
		switch v["type"] {
		case "text":
			str := html.EscapeString(data["text"].(string))
			str = strings.ReplaceAll(str, "\n", "<br/>")
			text += tmpl.HTML(str)
		case "image":
			// url := data["url"].(string)
			// text += tmpl.Get(config.GetTheme()).Image().SetSrc(tmpl.HTML(url)).GetContent()
			// text += tmpl.HTML(fmt.Sprintf(`[CQ:image,url=<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>]`, url, url))
			text += common.GetImageWithCache(data)
		case "face":
			face := v["data"].(global.MSG)
			if face["id"].(int32) <= 274 {
				text += tmpl.Get(config.GetTheme()).Image().SetSrc(tmpl.HTML(fmt.Sprintf("/admin/qq/faceimg/%d.gif", face["id"]))).
					SetHeight("24").
					SetWidth("24").
					GetContent()
			} else {
				text += tmpl.HTML(fmt.Sprintf("[CQ:face,id=%d]", face["id"]))
			}
		case "at":
			text += tmpl.HTML(fmt.Sprintf(`%s(%d)`, data["display"], data["target"]))
		default: // 其他消息类型
			d, _ := json.Marshal(v["data"])
			text += tmpl.HTML(fmt.Sprintf("[%s:%s]", v["type"], string(d)))
		}
	}
	components := tmpl.Get(config.GetTheme())
	return components.Row().SetContent(
		components.Col().SetSize(types.Size(11, 11, 11)).SetContent(
			tmpl.HTML(msgbox(msg.GetAttribute().SenderName, text, msg.GetAttribute().SenderUin == l.Cli.Uin)),
		).GetContent(),
	).GetContent()
}
func (l *Dologin) parseGuildChannelMsg(msg *db.StoredGuildChannelMessage) template.HTML {
	var text template.HTML
	for _, v := range msg.Content {
		data := v["data"].(global.MSG)
		switch v["type"] {
		case "text":
			str := html.EscapeString(data["text"].(string))
			str = strings.ReplaceAll(str, "\n", "<br/>")
			text += tmpl.HTML(str)
		case "image":
			// url := data["url"].(string)
			// text += tmpl.Get(config.GetTheme()).Image().SetSrc(tmpl.HTML(url)).GetContent()
			// text += tmpl.HTML(fmt.Sprintf(`[CQ:image,url=<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>]`, url, url))
			text += common.GetImageWithCache(data)
		case "face":
			face := v["data"].(global.MSG)
			if face["id"].(int32) <= 274 {
				text += tmpl.Get(config.GetTheme()).Image().SetSrc(tmpl.HTML(fmt.Sprintf("/admin/qq/faceimg/%d.gif", face["id"]))).
					SetHeight("24").
					SetWidth("24").
					GetContent()
			} else {
				text += tmpl.HTML(fmt.Sprintf("[CQ:face,id=%d]", face["id"]))
			}
		case "at":
			text += tmpl.HTML(fmt.Sprintf(`%s(%d)`, data["display"], data["target"]))
		default: // 其他消息类型
			d, _ := json.Marshal(v["data"])
			text += tmpl.HTML(fmt.Sprintf("[%s:%s]", v["type"], string(d)))
		}
	}
	components := tmpl.Get(config.GetTheme())
	return components.Row().SetContent(
		components.Col().SetSize(types.Size(11, 11, 11)).SetContent(
			tmpl.HTML(msgbox(msg.Attribute.SenderName, text, int64(msg.Attribute.SenderTinyID) == l.Cli.Uin)),
		).GetContent(),
	).GetContent()
}
func msgbox(name string, text template.HTML, isSelf bool) string {
	var box string
	if !isSelf {
		box = fmt.Sprintf(`
<div class="row" style="text-align:left;margin-left:15px;margin-right:15px;">
<div class="row">
%s<i class="fa fa-arrow-right"></i>
</div>
<p>%s</p>
</div>
`, name, text)
	} else {
		box = fmt.Sprintf(`
<div class="row" style="text-align:right;margin-left:15px;margin-right:15px;">
<div class="row">
<i class="fa fa-arrow-left"></i>%s
</div>
<p>%s</p>
</div>
`, name, text)
	}
	return box
}

// DeviceInfo 读取 device.json
func (l *Dologin) DeviceInfo(ctx iris.Context) (types.Panel, error) {
	if err := l.CheckQQlogin(ctx); err != nil {
		return types.Panel{}, nil
	}
	components := tmpl.Get(config.GetTheme())
	if ctx.Method() == "POST" { // post处理
		protol := ctx.PostValueIntDefault("protocol", 0)
		switch protol {
		case 1:
			client.SystemDeviceInfo.Protocol = client.AndroidPhone
		case 2:
			client.SystemDeviceInfo.Protocol = client.AndroidWatch
		case 3:
			client.SystemDeviceInfo.Protocol = client.MacOS
		case 4:
			client.SystemDeviceInfo.Protocol = client.QiDian
		case 5:
			client.SystemDeviceInfo.Protocol = client.IPad
		default:
			log.Error("protocol not changed")
			jsonStr := ctx.PostValue("json")
			if err := client.SystemDeviceInfo.ReadJson([]byte(jsonStr)); err != nil {
				return jump.Error(common.Msg{
					Msg: "设置失败了，json解析错误",
					URL: "/admin/qq/deviceinfo",
				}), nil
			}
		}
		_ = os.WriteFile("device.json", client.SystemDeviceInfo.ToJson(), 0o644)
		return jump.Success(common.Msg{
			Msg: "设置成功",
			URL: "/admin/qq/deviceinfo",
		}), nil
	}
	if err := client.SystemDeviceInfo.ReadJson([]byte(global.ReadAllText("device.json"))); err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("加载设备信息失败: %v", err),
			URL: "/admin/qq/info",
		}), nil
	}
	btn1 := components.Button().SetType("submit").
		SetContent("确认提交").
		SetThemePrimary().
		SetOrientationRight().
		SetLoadingText(icon.Icon("fa-spinner fa-spin", 2) + `Save`).
		GetContent()
	btn2 := components.Button().SetType("reset").
		SetContent("重置").
		SetThemeWarning().
		SetOrientationLeft().
		GetContent()
	col2 := components.Col().SetSize(types.SizeMD(8)).
		SetContent(btn1 + btn2).GetContent()
	var panel = types.NewFormPanel()
	panel.AddField("快速修改登录协议", "protocol", db2.Int, form.SelectSingle).
		FieldOptions(types.FieldOptions{
			{Text: "Unset", Value: fmt.Sprintf("%d", client.Unset)},
			{Text: "AndroidPhone", Value: fmt.Sprintf("%d", client.AndroidPhone)},
			{Text: "AndroidWatch", Value: fmt.Sprintf("%d", client.AndroidWatch)},
			{Text: "MacOS", Value: fmt.Sprintf("%d", client.MacOS)},
			{Text: "QiDian", Value: fmt.Sprintf("%d", client.QiDian)},
			{Text: "IPad", Value: fmt.Sprintf("%d", client.IPad)},
		}).FieldDefault(fmt.Sprintf("%d", client.SystemDeviceInfo.Protocol))
	panel.AddField("全局覆盖device.json", "json", db2.Varchar, form.TextArea).FieldPlaceholder("完整的device.json").FieldDefault(string(client.SystemDeviceInfo.ToJson())).FieldMust()
	panel.SetTabGroups(types.TabGroups{
		{"protocol", "json"},
	})
	panel.SetTabHeaders("调整你的device.json")
	fields, headers := panel.GroupField()
	aform := components.Form().
		SetTabHeaders(headers).
		SetTabContents(fields).
		SetPrefix(config.PrefixFixSlash()).
		SetUrl("/admin/qq/deviceinfo").
		SetOperationFooter(col2)
	return types.Panel{
		Content: components.Box().
			SetHeader(aform.GetDefaultBoxHeader(true)).
			WithHeadBorder().
			SetBody(aform.GetContent()).
			GetContent(),
		Title:       "调整你的device.json",
		Callbacks:   panel.Callbacks,
		Description: "可以快速修改登录类型，也可以全部json上传（全局覆盖的时候，协议选unset，否则只会变更登录协议)",
	}, nil
}
