package main

import (
	_ "github.com/GoAdminGroup/go-admin/adapter/iris"              // web framework adapter
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/sqlite" // sql driver
	_ "github.com/GoAdminGroup/themes/adminlte"                    // ui theme
	"github.com/Mrs4s/go-cqhttp/internal/base"

	"github.com/Mrs4s/go-cqhttp/cmd/gocq"
	iris_admin "github.com/Mrs4s/go-cqhttp/cmd/iris_admin"
	_ "github.com/Mrs4s/go-cqhttp/db/leveldb"   // leveldb
	_ "github.com/Mrs4s/go-cqhttp/modules/mime" // mime检查模块
	_ "github.com/Mrs4s/go-cqhttp/modules/silk" // silk编码模块
	// 其他模块
	// _ "github.com/Mrs4s/go-cqhttp/db/mongodb"    // mongodb 数据库支持
	// _ "github.com/Mrs4s/go-cqhttp/modules/pprof" // pprof 性能分析
)

func main() {
	if checkWebui() {
		iris_admin.StartServer()
	} else {
		gocq.Main()
	}
}

func checkWebui() bool {
	base.Parse()
	return base.WebUI
}
