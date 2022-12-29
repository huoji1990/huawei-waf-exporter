package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/region"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 配置文件结构体
type Data struct {
	Data []Host
}

type Host struct {
	Id       string `json:"id"`
	HostName string `json:"hostname"`
}

// WAF监控项数值结构体
type WafRequest struct {
	Requests        float64
	Qps_peak        float64
	Inbound_traffic float64
}

// WAFExporter结构体
type WafExporter struct {
	WafRequestDesc *prometheus.Desc
	WafQpsPeakDesc *prometheus.Desc
	WafInboundDesc *prometheus.Desc
}

// 华为云认证/WAF客户端/时间戳变量定义
var (
	ak   = <AK>
	sk   = <SK>
	auth = basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		Build()
	client = ces.NewCesClient(
		ces.CesClientBuilder().
			WithRegion(region.ValueOf("cn-north-4")).
			WithCredential(auth).
			Build())
)

var Config Data

// 读取配置文件
func init() {
	file, _ := os.Open("./conf/config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&Config)
	if err != nil {
		panic(err)
	}
}

// 获取WAF监控相关指标数据
func waf_request(host string) WafRequest {
	request := &model.BatchListMetricDataRequest{}
	var listDimensionsMetrics = []model.MetricsDimension{
		{
			Name:  "waf_instance_id",
			Value: host,
		},
	}
	var listMetricsbody = []model.MetricInfo{
		{
			Namespace:  "SYS.WAF",
			MetricName: "requests",
			Dimensions: listDimensionsMetrics,
		},
		{
			Namespace:  "SYS.WAF",
			MetricName: "qps_peak",
			Dimensions: listDimensionsMetrics,
		},
		{
			Namespace:  "SYS.WAF",
			MetricName: "inbound_traffic",
			Dimensions: listDimensionsMetrics,
		},
	}
	request.Body = &model.BatchListMetricDataRequestBody{
		To:      int64(time.Now().UnixMilli()),
		From:    int64(time.Now().UnixMilli() - 600000),
		Filter:  "max",
		Period:  "1",
		Metrics: listMetricsbody,
	}
	response, _ := client.BatchListMetricData(request)
	data := *response.Metrics
	waf_data := WafRequest{}
	if len(data[0].Datapoints) == 2 {
		waf_data.Requests = *data[0].Datapoints[1].Max
	} else {
		waf_data.Requests = *data[0].Datapoints[0].Max
	}
	if len(data[1].Datapoints) == 2 {
		waf_data.Qps_peak = *data[1].Datapoints[1].Max
	} else {
		waf_data.Qps_peak = *data[1].Datapoints[0].Max
	}
	if len(data[2].Datapoints) == 2 {
		waf_data.Inbound_traffic = *data[2].Datapoints[1].Max
	} else {
		waf_data.Inbound_traffic = *data[2].Datapoints[0].Max
	}
	return waf_data
}

// 初始化exporter
func NewExporter() *WafExporter {
	return &WafExporter{
		WafRequestDesc: prometheus.NewDesc("waf_huawei_requests", "Current number of waf requests", []string{"hostname"}, nil),
		WafQpsPeakDesc: prometheus.NewDesc("waf_huawei_qps_peak", "Current number of waf qps_peak", []string{"hostname"}, nil),
		WafInboundDesc: prometheus.NewDesc("waf_huawei_inbound", "Current number of waf inbound", []string{"hostname"}, nil),
	}
}

// 采集器Describe方法
func (e *WafExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.WafRequestDesc
	ch <- e.WafQpsPeakDesc
	ch <- e.WafInboundDesc
}

// 采集器Collect方法
func (e *WafExporter) Collect(ch chan<- prometheus.Metric) {
	for i := 0; i < len(Config.Data); i++ {
		date := waf_request(Config.Data[i].Id)
		ch <- prometheus.MustNewConstMetric(e.WafRequestDesc, prometheus.GaugeValue, date.Requests, Config.Data[i].HostName)
		ch <- prometheus.MustNewConstMetric(e.WafQpsPeakDesc, prometheus.GaugeValue, date.Qps_peak, Config.Data[i].HostName)
		ch <- prometheus.MustNewConstMetric(e.WafInboundDesc, prometheus.GaugeValue, date.Inbound_traffic, Config.Data[i].HostName)
	}
}

func main() {

	registry := prometheus.NewRegistry()

	exporter := NewExporter()

	registry.Register(exporter)

	r := gin.Default()
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	r.Run(":9091")
}
