# 编译命令
go build main.go

# Docker编译命令
docker build -t huawei-waf-exporter .

# Docker运行命令
docker run -p 9091:9091 huawei-waf-exporter

# 项目说明
该项目是将华为云的WAF的一些监控指标数据转换为promethues的数据格式，从而将华为云WAF的监控指标数据接入到promethues中，丰富更多的监控展示及告警。
