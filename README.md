# etcd_sd
etcd sd for prometheus，consul_sd_configs，file_sd_configs
> prometheus不支持etcd自动服务发现，因此写了这个脚本借道file_sd_configs实现基于tcd的自动服务发现

```
./etcd_sd -target-file /home/xiewj/container/iotmicro/monitor/prometheus/tgroups.json
```

生成的tgroups.json内容格式如下
```
[{"targets":["10.16.1.169:38659"],"labels":{"project_name":"alive-meterics"}}]
```

prometheus.yml
```
global:
  scrape_interval: 15s # 默认15s 全局每次数据收集的间隔
  evaluation_interval: 15s # 规则扫描时间间隔是15秒，默认不填写是 1分钟
  scrape_timeout: 5s    #超时时间
alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']

rule_files:
  - "/etc/prometheus/rule/*.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
  - job_name: 'node'
    scrape_interval: 8s
    static_configs:
      - targets: ['node-exporter:9100']
  - job_name: 'metrics-test'
    file_sd_configs:
      - files: 
        - /etc/prometheus/tgroups.json
        refresh_interval: 10s
```
