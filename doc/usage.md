# GO语言脚手架(gostarter)使用手册

## 安装使用

1. 安装`gogo`命令 `go get -u -v github.com/bingoohuang/`
1. 创建工程 `gogo -dir somepath/yourprojectname -disableCache`
1. 进入新工程目录 `cd somepath/yourprojectname`
1. 下载资源静态化工具 `go get -u -v github.com/bingoohuang/statiq`
1. 下载依赖资源，打包资源 `./gr.sh`
1. 编译项目 `/gb.py`
1. 运行项目 `yourprojectname -logrus=false -u`
1. 使用 goland 或 vscode 继续开发功能

## 脚手架包括功能列表

1. toml配置文件 及 viper 聚合配置使用
1. logrus 日志按天滚动生成
1. res 资源内嵌
1. http框架gin使用
1. ctl控制脚本支持(start/stop/restart/tail等)
1. reload supported by kill -USR2 pid
1. pprof 支持

    - `./gostarter --pprof-addr localhost:6060`
    - `open http://localhost:6060/debug/pprof in explorer`
    - or 可视化数据（火焰图），见如下：
    - `curl http://localhost:6060/debug/pprof/heap > heap.prof`
    - `go get -u github.com/google/pprof`
    - `pprof -http=:8080 heap.prof`
