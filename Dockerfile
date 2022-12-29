FROM golang:1.18.1 as huawei-waf-exporter

WORKDIR /go/src/huawei-waf-exporter

COPY . .

RUN export GO111MODULE=on && \
    export GOPROXY=https://goproxy.cn && \
    go build -o huawei-waf-exporter . && \
    tar cvf pack.tar huawei-waf-exporter conf/

FROM centos:7

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /huawei-waf-exporter

COPY --from=huawei-waf-exporter /go/src/huawei-waf-exporter/pack.tar /ks-waf-exporter

RUN tar -xvf pack.tar && rm -rf pack.tar

CMD ["./huawei-waf-exporter"]
