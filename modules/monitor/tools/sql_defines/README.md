```sh
# 目录名为 alert_type、文件名为 alert_index

# 生成 通知模版 SQL(遍历目录)
go run main.go sql --table=sp_alert_notify_template --dir=alerts --alert=true

# 生成 通知模版 SQL(单个文件)
go run main.go sql --table=sp_alert_notify_template --file=alerts/machine/machine_cpu.yml --alert=true

# 生成 告警计算模版 SQL(遍历目录)
go run main.go sql --table=sp_alert_rules --ext=json --dir=alerts --alert=true

# 生成 告警计算模版 SQL(单个文件)
go run main.go sql --table=sp_alert_rules --ext=json --file=alerts/machine/machine_cpu.json --alert=true

# 生成 系统内置 metrics 表达式 SQL
go run main.go sql --table=sp_metric_expression --ext=json --dir=metrics --alert=false

# 生成 钉钉通知的 dignding.http 给 vscode REST Client 插件使用
go run main.go http

```