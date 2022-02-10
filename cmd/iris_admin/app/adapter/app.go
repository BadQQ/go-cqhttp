package adapter

import (
	"errors"
	"fmt"
	"strings"

	config2 "github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	tmpl "github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/icon"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"github.com/kataras/iris/v12"
	"github.com/scjtqs2/bot_adapter/config"

	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/utils/common"
	"github.com/Mrs4s/go-cqhttp/cmd/iris_admin/utils/jump"
)

// GetAppList 获取bot-adapter的app列表
func (i *Info) GetAppList(ctx iris.Context) (types.Panel, error) {
	if err := CheckAuth(ctx); err != nil {
		return types.Panel{}, nil
	}
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/info"
	} else if ctx.GetReferrer().Path == "/admin/adapter/applist" {
		re = "/admin/qq/info"
	}
	fs := make([]map[string]types.InfoItem, 0)
	comp := tmpl.Get(config2.GetTheme())
	linkAdd := comp.Link().SetClass("btn btn-sm btn-info btn-flat pull-right").SetContent("添加app").SetURL("/admin/adapter/addapp").GetContent()
	conf, err := i.getConfig()
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取bot-adapter的配置信息失败:%s", err.Error()),
			URL: re,
		}), nil
	}
	for _, g := range conf.Plugins {
		linkView := comp.Link().SetClass("btn btn-sm btn-danger").SetContent("查看").SetURL("/admin/adapter/viewapp?appid=" + g.AppID).GetContent()
		linkEdit := comp.Link().SetClass("btn btn-sm btn-primary").SetContent("编辑").SetURL("/admin/adapter/editapp?appid=" + g.AppID).GetContent()
		linkDel := comp.Link().SetClass("btn btn-sm btn-default").SetContent("删除").SetURL("/admin/adapter/delapp?appid=" + g.AppID).GetContent()
		fs = append(fs, map[string]types.InfoItem{
			"appID":     {Content: tmpl.HTML(g.AppID)},
			"appSecret": {Content: tmpl.HTML(g.AppSecret)},
			"appName":   {Content: tmpl.HTML(g.PluginName)},
			"操作":        {Content: linkView + linkEdit + linkDel},
		})
	}
	param := parameter.GetParam(ctx.Request().URL, 100)
	table := comp.Table().
		SetInfoList(common.PageSlice(fs, param.PageInt, param.PageSizeInt)).
		SetThead(types.Thead{
			{Head: "appID", Width: "15%"},
			{Head: "appSecret", Width: "30%"},
			{Head: "appName", Width: "30%"},
			{Head: "操作"},
		}).SetMinWidth("0.01%")

	body := table.GetContent()
	linkBack := comp.Link().SetURL(re).SetContent("返回").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	return types.Panel{
		Content: comp.Box().
			SetHeader(linkAdd).
			SetBody(body).
			SetFooter(linkBack).
			SetNoPadding().
			WithHeadBorder().
			GetContent(),
		Title:       "应用列表",
		Description: tmpl.HTML("bot-adapter的应用列表"),
	}, nil
}

// ViewApp 查看app信息
func (i *Info) ViewApp(ctx iris.Context) (types.Panel, error) {
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
	comp := tmpl.Get(config2.GetTheme())
	i.AdapterConf, err = i.getConfig()
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取bot-adapter的配置信息失败:%s", err.Error()),
			URL: re,
		}), nil
	}
	appid := ctx.URLParam("appid")
	info, err := i.getPluginInfoFromConf(appid)
	if err != nil {
		return jump.Error(common.Msg{
			Msg: err.Error(),
			URL: re,
		}), nil
	}
	permissing, err := i.getPluginPermission(appid)
	if err != nil {
		//  默认权限配置
		permissing = &config.PermissionConfig{
			AppID:        appid,
			IsAdmin:      false,
			IsOnlyCqhttp: true,
		}
	}
	getRegisterKeyWords := func(msgType string) []string {
		var ret []string
		for _, word := range info.RegisterKeyWords {
			if word.MsgType == msgType {
				return word.PrefixKeyWords
			}
		}
		return ret
	}
	var panel = types.NewFormPanel()
	panel.AddField("AppName", "plugin_name", db.Varchar, form.Text).FieldDefault(info.PluginName).FieldMust()
	panel.AddField("App描述", "plugin_desc", db.Varchar, form.TextArea).FieldDefault(info.PluginDesc)
	panel.AddField("AppID", "app_id", db.Varchar, form.Text).FieldDefault(info.AppID).FieldMust()
	panel.AddField("AppSecret", "app_secret", db.Varchar, form.Text).FieldDefault(info.AppSecret)
	panel.AddField("POST推送地址", "post_addr", db.Varchar, form.Text).FieldDefault(info.PostAddr)
	panel.AddField("推送秘钥（encrypt_key）", "encrypt_key", db.Varchar, form.Text).FieldDefault(info.EncryptKey)
	private := getRegisterKeyWords("private")
	group := getRegisterKeyWords("group")
	panel.AddField("私聊消息拦截前缀（一行一个）", "resigter_private", db.Text, form.TextArea).FieldDefault(func(p []string) string {
		var str string
		for _, s := range p {
			str += fmt.Sprintf("%s\r\n", s)
		}
		return str
	}(private))
	panel.AddField("群消息拦截前缀（一行一个）", "resigter_group", db.Text, form.TextArea).FieldDefault(func(p []string) string {
		var str string
		for _, s := range p {
			str += fmt.Sprintf("%s\r\n", s)
		}
		return str
	}(group))
	panel.AddField("是否管理权限", "permission_is_admin", db.Tinyint, form.Radio).
		FieldOptions(types.FieldOptions{
			{Text: "是", Value: "1"},
			{Text: "否", Value: "0"},
		}).FieldDisplay(func(value types.FieldModel) interface{} {
		if permissing.IsAdmin {
			return "1"
		}
		return "0"
	})
	panel.AddField("是否CQhttp基本权限", "permission_is_only_cqhttp", db.Tinyint, form.Radio).
		FieldOptions(types.FieldOptions{
			{Text: "是", Value: "1"},
			{Text: "否", Value: "0"},
		}).FieldDisplay(func(value types.FieldModel) interface{} {
		if permissing.IsOnlyCqhttp {
			return "1"
		}
		return "0"
	})
	panel.AddField("推送权限", "permission_push_permissions", db.Varchar, form.SelectBox).
		FieldOptions(i.getPushPermissionsOptions()).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return permissing.PushPermissions
		})
	panel.AddField("api权限", "permission_api_permissions", db.Varchar, form.SelectBox).
		FieldOptions(i.getAPIPermissionOptions()).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return permissing.ApiPermissions
		})
	panel.SetTabGroups(types.TabGroups{
		{"plugin_name", "app_id", "app_secret", "post_addr", "encrypt_key", "resigter_private", "resigter_group", "permission_is_admin", "permission_is_only_cqhttp", "permission_push_permissions", "permission_api_permissions"},
	})
	panel.SetTabHeaders("app的详细信息")

	fields, headers := panel.GroupField()
	aform := comp.Form().
		SetTabHeaders(headers).
		SetTabContents(fields).
		SetUrl("/admin/adapter/saveapp"). // 设置表单请求路由
		SetTitle("Form").
		SetOperationFooter("").
		GetContent()
	body := aform
	linkBack := comp.Link().SetURL(re).SetContent("返回").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetFooter(linkBack).
			SetNoPadding().
			WithHeadBorder().
			GetContent(),
		Title:       "应用详情",
		Description: tmpl.HTML("bot-adapter的应用详情 <br> appid不可重复。<br> POST推送地址不需要留空即可 <br> 填了POST推送地址，推送秘钥必须填上"),
	}, nil
}

// EditApp 编辑app
func (i *Info) EditApp(ctx iris.Context) (types.Panel, error) {
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
	comp := tmpl.Get(config2.GetTheme())
	i.AdapterConf, err = i.getConfig()
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取bot-adapter的配置信息失败:%s", err.Error()),
			URL: re,
		}), nil
	}
	appid := ctx.URLParam("appid")
	info, err := i.getPluginInfoFromConf(appid)
	if err != nil {
		return jump.Error(common.Msg{
			Msg: err.Error(),
			URL: re,
		}), nil
	}
	permissing, err := i.getPluginPermission(appid)
	if err != nil {
		//  默认权限配置
		permissing = &config.PermissionConfig{
			AppID:        appid,
			IsAdmin:      false,
			IsOnlyCqhttp: true,
		}
	}
	getRegisterKeyWords := func(msgType string) []string {
		var ret []string
		for _, word := range info.RegisterKeyWords {
			if word.MsgType == msgType {
				return word.PrefixKeyWords
			}
		}
		return ret
	}
	var panel = types.NewFormPanel()
	panel.AddField("App名称", "plugin_name", db.Varchar, form.Text).FieldDefault(info.PluginName).FieldMust()
	panel.AddField("App描述", "plugin_desc", db.Varchar, form.TextArea).FieldDefault(info.PluginDesc)
	panel.AddField("AppID", "app_id", db.Varchar, form.Text).FieldDefault(info.AppID).FieldMust()
	panel.AddField("AppSecret", "app_secret", db.Varchar, form.Text).FieldDefault(info.AppSecret).FieldMust()
	panel.AddField("POST推送地址", "post_addr", db.Varchar, form.Text).FieldDefault(info.PostAddr)
	panel.AddField("推送秘钥（encrypt_key）", "encrypt_key", db.Varchar, form.Text).FieldDefault(info.EncryptKey)
	private := getRegisterKeyWords("private")
	group := getRegisterKeyWords("group")
	panel.AddField("私聊消息拦截前缀（一行一个）", "resigter_private", db.Text, form.TextArea).FieldDefault(func(p []string) string {
		var str string
		for _, s := range p {
			str += fmt.Sprintf("%s\r\n", s)
		}
		return str
	}(private))
	panel.AddField("群消息拦截前缀（一行一个）", "resigter_group", db.Text, form.TextArea).FieldDefault(func(p []string) string {
		var str string
		for _, s := range p {
			str += fmt.Sprintf("%s\r\n", s)
		}
		return str
	}(group))
	panel.AddField("是否管理权限", "permission_is_admin", db.Tinyint, form.Radio).
		FieldOptions(types.FieldOptions{
			{Text: "是", Value: "1"},
			{Text: "否", Value: "0"},
		}).FieldDisplay(func(value types.FieldModel) interface{} {
		if permissing.IsAdmin {
			return "1"
		}
		return "0"
	}).FieldMust()
	panel.AddField("是否CQhttp基本权限", "permission_is_only_cqhttp", db.Tinyint, form.Radio).
		FieldOptions(types.FieldOptions{
			{Text: "是", Value: "1"},
			{Text: "否", Value: "0"},
		}).FieldDisplay(func(value types.FieldModel) interface{} {
		if permissing.IsOnlyCqhttp {
			return "1"
		}
		return "0"
	}).FieldMust()
	panel.AddField("推送权限", "permission_push_permissions", db.Varchar, form.SelectBox).
		FieldOptions(i.getPushPermissionsOptions()).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return permissing.PushPermissions
		})
	panel.AddField("api权限", "permission_api_permissions", db.Varchar, form.SelectBox).
		FieldOptions(i.getAPIPermissionOptions()).
		FieldDisplay(func(value types.FieldModel) interface{} {
			return permissing.ApiPermissions
		})
	panel.SetTabGroups(types.TabGroups{
		{"plugin_name", "app_id", "app_secret", "post_addr", "encrypt_key", "resigter_private", "resigter_group", "permission_is_admin", "permission_is_only_cqhttp", "permission_push_permissions", "permission_api_permissions"},
	})
	panel.SetTabHeaders("app的详细信息")

	fields, headers := panel.GroupField()
	btn1 := comp.Button().SetType("submit").
		SetContent("确认提交").
		SetThemePrimary().
		SetOrientationRight().
		SetLoadingText(icon.Icon("fa-spinner fa-spin", 2) + `Save`).
		GetContent()
	btn2 := comp.Button().SetType("reset").
		SetContent("重置").
		SetThemeWarning().
		SetOrientationLeft().
		GetContent()
	col1 := comp.Col().SetSize(types.SizeMD(8)).
		SetContent(btn1 + btn2).GetContent()
	aform := comp.Form().
		SetTabHeaders(headers).
		SetTabContents(fields).
		SetHiddenFields(map[string]string{
			"action": "update",
		}).
		SetUrl("/admin/adapter/saveapp"). // 设置表单请求路由
		SetTitle("Form").
		SetOperationFooter(col1).
		GetContent()
	body := aform
	linkBack := comp.Link().SetURL(re).SetContent("返回").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetFooter(linkBack).
			SetNoPadding().
			WithHeadBorder().
			GetContent(),
		Title:       "应用详情",
		Description: tmpl.HTML("bot-adapter的应用详情 <br> appid不可重复。<br> POST推送地址不需要留空即可 <br> 填了POST推送地址，推送秘钥必须填上"),
	}, nil
}

// DelApp 删除app
func (i *Info) DelApp(ctx iris.Context) (types.Panel, error) {
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
	i.AdapterConf, err = i.getConfig()
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取bot-adapter的配置信息失败:%s", err.Error()),
			URL: re,
		}), nil
	}
	appid := ctx.URLParam("appid")
	if appid == "" {
		return jump.Error(common.Msg{
			Msg: "empty appid",
			URL: re,
		}), nil
	}
	newPlugins := make([]*config.PluginConfig, 0)
	newPermission := make([]*config.PermissionConfig, 0)
	for _, plugin := range i.AdapterConf.Plugins {
		if plugin.AppID == appid {
			continue
		}
		newPlugins = append(newPlugins, plugin)
	}
	for _, permission := range i.AdapterConf.Permissions {
		if permission.AppID == appid {
			continue
		}
		newPermission = append(newPermission, permission)
	}
	i.AdapterConf.Permissions = newPermission
	i.AdapterConf.Plugins = newPlugins
	_, err = i.saveAdapterConfig(i.AdapterConf)
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

// AddApp 新增app
func (i *Info) AddApp(ctx iris.Context) (types.Panel, error) {
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
	comp := tmpl.Get(config2.GetTheme())
	i.AdapterConf, err = i.getConfig()
	if err != nil {
		return jump.Error(common.Msg{
			Msg: fmt.Sprintf("获取bot-adapter的配置信息失败:%s", err.Error()),
			URL: re,
		}), nil
	}
	var panel = types.NewFormPanel()
	panel.AddField("App名称", "plugin_name", db.Varchar, form.Text).FieldMust()
	panel.AddField("App描述", "plugin_desc", db.Varchar, form.TextArea)
	panel.AddField("AppID", "app_id", db.Varchar, form.Text).FieldMust()
	panel.AddField("AppSecret", "app_secret", db.Varchar, form.Text).FieldMust()
	panel.AddField("POST推送地址", "post_addr", db.Varchar, form.Text)
	panel.AddField("推送秘钥（encrypt_key）", "encrypt_key", db.Varchar, form.Text)
	panel.AddField("私聊消息拦截前缀（一行一个）", "resigter_private", db.Text, form.TextArea)
	panel.AddField("群消息拦截前缀（一行一个）", "resigter_group", db.Text, form.TextArea)
	panel.AddField("是否管理权限", "permission_is_admin", db.Tinyint, form.Radio).
		FieldOptions(types.FieldOptions{
			{Text: "是", Value: "1"},
			{Text: "否", Value: "0"},
		}).FieldMust()
	panel.AddField("是否CQhttp基本权限", "permission_is_only_cqhttp", db.Tinyint, form.Radio).
		FieldOptions(types.FieldOptions{
			{Text: "是", Value: "1"},
			{Text: "否", Value: "0"},
		}).FieldMust()
	panel.AddField("推送权限", "permission_push_permissions", db.Varchar, form.SelectBox).
		FieldOptions(i.getPushPermissionsOptions())
	panel.AddField("api权限", "permission_api_permissions", db.Varchar, form.SelectBox).
		FieldOptions(i.getAPIPermissionOptions())
	panel.SetTabGroups(types.TabGroups{
		{"plugin_name", "app_id", "app_secret", "post_addr", "encrypt_key", "resigter_private", "resigter_group", "permission_is_admin", "permission_is_only_cqhttp", "permission_push_permissions", "permission_api_permissions"},
	})
	panel.SetTabHeaders("新增App")

	fields, headers := panel.GroupField()
	btn1 := comp.Button().SetType("submit").
		SetContent("确认提交").
		SetThemePrimary().
		SetOrientationRight().
		SetLoadingText(icon.Icon("fa-spinner fa-spin", 2) + `Save`).
		GetContent()
	btn2 := comp.Button().SetType("reset").
		SetContent("重置").
		SetThemeWarning().
		SetOrientationLeft().
		GetContent()
	col1 := comp.Col().SetSize(types.SizeMD(8)).
		SetContent(btn1 + btn2).GetContent()
	aform := comp.Form().
		SetTabHeaders(headers).
		SetTabContents(fields).
		SetHiddenFields(map[string]string{
			"action": "add",
		}).
		SetUrl("/admin/adapter/saveapp"). // 设置表单请求路由
		SetTitle("Form").
		SetOperationFooter(col1).
		GetContent()
	body := aform
	linkBack := comp.Link().SetURL(re).SetContent("返回").SetClass("btn btn-sm btn-info btn-flat pull-left").GetContent()
	return types.Panel{
		Content: comp.Box().
			SetBody(body).
			SetFooter(linkBack).
			SetNoPadding().
			WithHeadBorder().
			GetContent(),
		Title:       "新建app",
		Description: tmpl.HTML("bot-adapter 新增app <br> appid不可重复。<br> POST推送地址不需要留空即可 <br> 填了POST推送地址，推送秘钥必须填上"),
	}, nil
}

// SaveApp 保存app信息
func (i *Info) SaveApp(ctx iris.Context) (types.Panel, error) {
	appID := ctx.PostValueTrim("app_id")           // 不能空
	pluginName := ctx.PostValueTrim("plugin_name") // 不能空
	pluginDesc := ctx.PostValueTrim("plugin_desc")
	appSecret := ctx.PostValueTrim("app_secret") // 不能空
	postAddr := ctx.PostValueTrim("post_addr")
	encryptKye := ctx.PostValueTrim("encrypt_key")
	resigterPrivate := ctx.PostValueTrim("resigter_private")
	resigterGroup := ctx.PostValueTrim("resigter_group")
	permissionIsAdmin := ctx.PostValueIntDefault("permission_is_admin", 0)
	permissionIsOnlyCqhttp := ctx.PostValueIntDefault("permission_is_only_cqhttp", 0)
	permissionPushPermissions := ctx.PostValues("permission_push_permissions[]")
	permissionAPIPermissions := ctx.PostValues("permission_api_permissions[]")
	re := ctx.GetReferrer().URL
	if re == "" {
		re = "/admin/qq/info"
	} else if ctx.GetReferrer().Path == "/admin/adapter/saveapp" {
		re = "/admin/qq/info"
	}
	if appID == "" {
		return jump.Error(common.Msg{
			Msg: "app_id不能为空",
			URL: re,
		}), nil
	}
	if pluginName == "" {
		return jump.Error(common.Msg{
			Msg: "app_name不能为空",
			URL: re,
		}), nil
	}

	if appSecret == "" {
		return jump.Error(common.Msg{
			Msg: "app_secret不能为空",
			URL: re,
		}), nil
	}
	var info *config.PluginConfig
	switch ctx.PostValue("action") {
	case "add":
		// 校验appid是否已存在
		_, err := i.getPluginInfoFromConf(appID)
		if err == nil {
			return jump.Error(common.Msg{
				Msg: "app_id已存在",
				URL: re,
			}), nil
		}
	case "update":
	default:
		return types.Panel{}, nil
	}
	info = &config.PluginConfig{
		AppID:      appID,
		AppSecret:  appSecret,
		EncryptKey: encryptKye,
		PostAddr:   postAddr,
		PluginName: pluginName,
		PluginDesc: pluginDesc,
		RegisterKeyWords: func(private, group string) []*config.RegisterKeyWords {
			var keys []*config.RegisterKeyWords
			if private != "" {
				var p []string
				// reg, _ := regexp.Compile(`<p>(.*?)</p>`)
				// for _, s := range reg.FindAllStringSubmatch(private, -1) {
				// 	p = append(p, strings.TrimSpace(s[1]))
				// }
				private = strings.ReplaceAll(private, "\r\n", "\n")
				private = strings.ReplaceAll(private, "\r", "\n")
				for _, s := range strings.Split(private, "\n") {
					p = append(p, strings.TrimSpace(s))
				}
				keys = append(keys, &config.RegisterKeyWords{
					MsgType:        "private",
					PrefixKeyWords: p,
				})
			}
			if group != "" {
				var g []string
				// reg, _ := regexp.Compile(`<p>(.*?)</p>`)
				// for _, s := range reg.FindAllStringSubmatch(group, -1) {
				// 	g = append(g, strings.TrimSpace(s[1]))
				// }
				group = strings.ReplaceAll(group, "\r\n", "\n")
				group = strings.ReplaceAll(group, "\r", "\n")
				for _, s := range strings.Split(group, "\n") {
					g = append(g, strings.TrimSpace(s))
				}
				keys = append(keys, &config.RegisterKeyWords{
					MsgType:        "group",
					PrefixKeyWords: g,
				})
			}
			return keys
		}(resigterPrivate, resigterGroup),
	}
	// 更新当前缓存的config
	updatePlugsins := func(info *config.PluginConfig) {
		var plugins []*config.PluginConfig
		isExit := false
		for _, plugin := range i.AdapterConf.Plugins {
			if plugin.AppID == info.AppID {
				plugins = append(plugins, info)
				isExit = true
				continue
			}
			plugins = append(plugins, plugin)
		}
		if !isExit {
			plugins = append(plugins, info)
		}
		i.AdapterConf.Plugins = plugins
	}
	updatePermission := func(info *config.PermissionConfig) {
		var permissions []*config.PermissionConfig
		isExit := false
		for _, permission := range i.AdapterConf.Permissions {
			if permission.AppID == info.AppID {
				permissions = append(permissions, info)
				isExit = true
				continue
			}
			permissions = append(permissions, permission)
		}
		if !isExit {
			permissions = append(permissions, info)
		}
		i.AdapterConf.Permissions = permissions
	}
	permission := &config.PermissionConfig{
		AppID:           appID,
		IsAdmin:         permissionIsAdmin == 1,
		IsOnlyCqhttp:    permissionIsOnlyCqhttp == 1,
		PushPermissions: permissionPushPermissions,
		ApiPermissions:  permissionAPIPermissions,
	}
	updatePermission(permission)
	updatePlugsins(info)
	_, err := i.saveAdapterConfig(i.AdapterConf)
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

// getPluginInfoFromConf 根据appID拉取 app信息
func (i *Info) getPluginInfoFromConf(appid string) (conf *config.PluginConfig, err error) {
	if i.AdapterConf == nil {
		i.AdapterConf, err = i.getConfig()
		if err != nil {
			return nil, err
		}
	}
	for _, plugin := range i.AdapterConf.Plugins {
		if plugin.AppID == appid {
			return plugin, nil
		}
	}
	return nil, errors.New("appId not found")
}

// getPluginPermission 根据appID 拉取配置信息中的数据
func (i *Info) getPluginPermission(appid string) (conf *config.PermissionConfig, err error) {
	if i.AdapterConf == nil {
		i.AdapterConf, err = i.getConfig()
		if err != nil {
			return nil, err
		}
	}
	for _, permission := range i.AdapterConf.Permissions {
		if permission.AppID == appid {
			return permission, nil
		}
	}
	return nil, errors.New("appId not found")
}

// getPushPermissionsOptions 生成推送权限列表
func (i *Info) getPushPermissionsOptions() types.FieldOptions {
	return types.FieldOptions{
		{Text: "私聊消息(message_private)", Value: config.PUSH_PERMISSION_MESSAGE_PRIVATE},
		{Text: "群消息(message_group)", Value: config.PUSH_PERMISSION_MESSAGE_GROUP},
		{Text: "群文件上传通知(notice_group_upload)", Value: config.PUSH_PERMISSION_NOTICE_GROUP_UPLOAD},
		{Text: "管理员变动(notice_group_admin)", Value: config.PUSH_PERMISSION_NOTICE_GROUP_ADMIN},
		{Text: "群成员减少(notice_group_decrease)", Value: config.PUSH_PERMISSION_NOTICE_GROUP_DECREASE},
		{Text: "群成功增加(notice_group_decrease)", Value: config.PUSH_PERMISSION_NOTICE_GROUP_INCREASE},
		{Text: "群成员禁言(notice_group_ban)", Value: config.PUSH_PERMISSION_NOTICE_GROUP_BAN},
		{Text: "加好友通知(notice_friend_add)", Value: config.PUSH_PERMISSION_NOTICE_FRIEND_ADD},
		{Text: "群消息撤回(notice_group_recall)", Value: config.PUSH_PERMISSION_NOTICE_GROUP_RECALL},
		{Text: "私聊消息撤回(notice_friend_recall)", Value: config.PUSH_PERMISSION_NOTICE_FRIEND_RECALL},
		{Text: "群内戳一戳(notice_notify_poke)", Value: config.PUSH_PERMISSION_NOTICE_NOTIFY_POKE},
		{Text: "群红包运气王(notice_notify_lucky_king)", Value: config.PUSH_PERMISSION_NOTICE_NOTIFY_LUCKY_KING},
		{Text: "群成员荣誉变更(notice_notify_honor)", Value: config.PUSH_PERMISSION_NOTICE_NOTIFY_HONOR},
		{Text: "添加好友请求(request_friend)", Value: config.PUSH_PERMISSION_REQUEST_FRIEND},
		{Text: "加群请求(request_group)", Value: config.PUSH_PERMISSION_REQUEST_GROUP},
		{Text: "生命周期(meta_event_lifecycle)", Value: config.PUSH_PERMISSION_META_EVENT_LIFECYCLE},
		{Text: "心跳(meta_event_heartbeat)", Value: config.PUSH_PERMISSION_META_EVENT_HEARTBEAT},
		{Text: "群成员名片更新(custom_notice_group_card)", Value: config.PUSH_CUSTOM_PERMISSION_NOTICE_GROUP_CARD},
		{Text: "接收离线文件(custom_notice_offline_file)", Value: config.PUSH_CUSTOM_PERMISSION_NOTICE_OFFLINE_FILE},
	}
}

// getAPIPermissionOptions 生成api权限列表
func (i *Info) getAPIPermissionOptions() types.FieldOptions {
	return types.FieldOptions{
		{Text: "发送私聊消息(send_private_msg)", Value: config.PERMISSION_FOR_SEND_PRIVATE_MSG},
		{Text: "发送群聊消息(send_group_msg)", Value: config.PERMISSION_FOR_SEND_GROUP_MSG},
		{Text: "发送消息(send_msg)", Value: config.PERMISSION_FOR_SEND_MSG},
		{Text: "撤回消息(delete_msg)", Value: config.PERMISSION_FOR_DELETE_MSG},
		{Text: "获取消息(get_msg)", Value: config.PERMISSION_FOR_GET_MSG},
		{Text: "获取合并转发消息(get_forward_msg)", Value: config.PERMISSION_FOR_GET_FORWARD_MSG},
		{Text: "发送好友赞(send_like)", Value: config.PERMISSION_FOR_SEND_LIKE},
		{Text: "群组踢人(set_group_kick)", Value: config.PERMISSION_FOR_SET_GROUP_KICK},
		{Text: "群组单人禁言(set_group_ban)", Value: config.PERMISSION_FOR_SET_GROUP_BAN},
		{Text: "群组匿名用户禁言(set_group_anonymous_ban)", Value: config.PERMISSION_FOR_SET_GROUP_ANONYMOUS_BAN},
		{Text: "群组全员禁言(set_group_group_whole_ban)", Value: config.PERMISSION_FOR_SET_GROUP_WHOLE_BAN},
		{Text: "群组设置管理员(set_group_admin)", Value: config.PERMISSION_FOR_SET_GROUP_ADMIN},
		{Text: "群组匿名(set_group_anonymous)", Value: config.PERMISSION_FOR_SET_GROUP_ANONYMOUS},
		{Text: "设置群名片(set_group_card)", Value: config.PERMISSION_FOR_SET_GROUP_CARD},
		{Text: "设置群名(set_group_name)", Value: config.PERMISSION_FOR_SET_GROUP_NAME},
		{Text: "退出群组(set_group_leave)", Value: config.PERMISSION_FOR_SET_GROUP_LEAVE},
		{Text: "设置群组专属头衔(set_group_special_title)", Value: config.PERMISSION_FOR_SET_GROUP_SPECIAL_TITLE},
		{Text: "处理好友请求(set_friend_add_request)", Value: config.PERMISSION_FOR_SET_FRIEND_ADD_REQUEST},
		{Text: "处理加群请求/邀请(set_group_add_request)", Value: config.PERMISSION_FOR_SET_GROUP_ADD_REQUEST},
		{Text: "获取登录号信息(get_login_info)", Value: config.PERMISSION_FOR_GET_LOGIN_INFO},
		{Text: "获取陌生人信息(get_stranger_info)", Value: config.PERMISSION_FOR_GET_STRANGER_INFO},
		{Text: "获取好友列表(get_friend_list)", Value: config.PERMISSION_FOR_GET_FRIEND_LIST},
		{Text: "获取群信息(get_group_info)", Value: config.PERMISSION_FOR_GET_GROUP_INFO},
		{Text: "获取群列表(get_group_list)", Value: config.PERMISSION_FOR_GET_GROUP_LIST},
		{Text: "获取群成员信息(get_group_member_info)", Value: config.PERMISSION_FOR_GET_GROUP_MEMBER_INFO},
		{Text: "获取群成员列表(get_group_member_list)", Value: config.PERMISSION_FOR_GET_GROUP_MEMBER_LIST},
		{Text: "获取群荣誉信息(get_group_honor_info)", Value: config.PERMISSION_FOR_GET_GROUP_HONOR_INFO},
		{Text: "获取Cookies(get_cookies)", Value: config.PERMISSION_FOR_GET_COOKIES},
		{Text: "获取 CSRF Token(get_csrf_token)", Value: config.PERMISSION_FOR_GET_CSRF_TOKEN},
		{Text: "获取 QQ 相关接口凭证(get_credentials)", Value: config.PERMISSION_FOR_SET_GROUP_ADD_REQUEST},
		{Text: "获取语音(get_record)", Value: config.PERMISSION_FOR_GET_RECORD},
		{Text: "获取图片(get_image)", Value: config.PERMISSION_FOR_GET_IMAGE},
		{Text: "检查是否可以发送图片(can_send_image)", Value: config.PERMISSION_FOR_CAN_SEND_IMAGE},
		{Text: "检查是否可以发送语音(can_send_record)", Value: config.PERMISSION_FOR_CAN_SEND_RECORD},
		{Text: "获取运行状态(get_status)", Value: config.PERMISSION_FOR_GET_STATUS},
		{Text: "获取版本信息(get_version_info)", Value: config.PERMISSION_FOR_GET_VERSION_INFO},
		{Text: "重启 OneBot 实现(set_restart)", Value: config.PERMISSION_FOR_SET_RESTART},
		{Text: "清理缓存(clean_cache)", Value: config.PERMISSION_FOR_CLEAN_CACHE},
		// go-cqhttp实现的一些非 onebot-11 的接口 以及于 onebot标准略微差异的API
		{Text: "设置群头像(custom_set_group_portrait)", Value: config.PERMISSION_FOR_CUSTOM_SET_GROUP_PROTRAIT},
		{Text: "获取中文分词(custom_get_word_slices)", Value: config.PERMISSION_FOR_CUSTOM_GET_WORD_SLICES},
		{Text: "图片OCR(custom_ocr_image)", Value: config.PERMISSION_FOR_CUSTOM_OCR_IMAGE},
		{Text: "获取群系统消息(custom_get_group_system_msg)", Value: config.PERMISSION_FOR_CUSTOM_GET_GROUP_SYSTEM_MSG},
		{Text: "上传群文件(custom_upload_group_file)", Value: config.PERMISSION_FOR_CUSTOM_UPLOAD_GROUP_FILE},
		{Text: "获取群@全体成员 剩余次数(custom_get_group_at_all_remain)", Value: config.PERMISSION_FOR_CUSTOM_GET_GROUP_AT_ALL_REMAIN},
		{Text: "获取VIP信息(custom_get_vip_info)", Value: config.PERMISSION_FOR_CUSTOM_GET_VIP_INFO},
		{Text: "发送群公告(custom_send_group_notice)", Value: config.PERMISSION_FOR_CUSTOM_SEND_GROUP_NOTICE},
		{Text: "重载事件过滤器(custom_reload_event_filter)", Value: config.PERMISSION_FOR_RELOAD_EVENT_FILTER},
		{Text: "发送合并转发（群）(custom_send_group_forward_msg)", Value: config.PERMISSION_FOR_CUSTOM_SEND_GROUP_FORWARD_msg},
		{Text: "获取群文件系统信息(custom_get_group_file_system_info)", Value: config.PERMISSION_FOR_CUSTOM_GET_GROUP_FILE_SYSTEM_INFO},
		{Text: "获取群根目录文件列表(custom_get_group_root_files)", Value: config.PERMISSION_FOR_CUSTOM_GET_GROUP_ROOT_FILES},
		{Text: "获取群子目录文件列表(custom_get_group_files_by_folder)", Value: config.PERMISSION_FOR_CUSTOM_GET_GROUP_FILES_BY_FOLDER},
		{Text: "获取群文件资源连接(custom_get_group_file_url)", Value: config.PERMISSION_FOR_CUSTOM_GET_GROUP_FILE_URL},
		{Text: "下载文件到缓存目录(custom_download_file)", Value: config.PERMISSION_FOR_CUSTOM_DOWNLOAD_FILE},
		{Text: "获取当前账号在线客户端列表(custom_get_online_clients)", Value: config.PERMISSION_FOR_CUSTOM_GET_ONLINE_CLIENTS},
		{Text: "获取群消息历史记录(custom_get_group_msg_history)", Value: config.PERMISSION_FOR_CUSTOM_GET_GROUP_MSG_HISTORY},
		{Text: "设置精华消息(custom_set_essence_msg)", Value: config.PERMISSION_FOR_CUSTOM_SET_ESSENCE_MSG},
		{Text: "移出精华消息(custom_delete_essence_msg)", Value: config.PERMISSION_FOR_CUSTOM_DELETE_ESSENCE_MSG},
		{Text: "获取精华消息列表(custom_get_essence_msg_list)", Value: config.PERMISSION_FOR_CUSTOM_GET_ESSENCE_MSG_LIST},
		{Text: "检查链接安全性(custom_check_url_safely)", Value: config.PERMISSION_FOR_CUSTOM_CHECK_URL_SAFELY},
		{Text: "获取在线机型(custom_get_model_show)", Value: config.PERMISSION_FOR_CUSTOM_GET_MODEL_SHOW},
		{Text: "设置在线机型(custom_set_model_show)", Value: config.PERMISSION_FOR_CUSTOM_SET_MODEL_SHOW},
		// 其他权限
		{Text: "拉取 bot_adapter 的配置文件(get_config)", Value: config.PERMISSION_FOR_GET_CONFIG},
		{Text: "直接更新覆盖 bot_adapter的配置(update_config)", Value: config.PERMISSION_FOR_UPDATE_CONFIG},
		{Text: "用来重启 bot_adapter的docekr(set_bot_adapter_kill)", Value: config.PERMISSION_FOR_SET_BOT_ADAPTER_KILL},
	}
}
