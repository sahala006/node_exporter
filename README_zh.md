[//]: # (<div align="left"><img src="docs/images/funasr_logo.jpg" width="400"/></div>)

(简体中文|[English](./README.md))

<a name="introduction"></a>
## 功能介绍
基于[node_exporter v1.9.0](https://github.com/prometheus/node_exporter/tree/v1.9.0 "node_exporter v1.9.0")来进行二次开发
- 定期上报主机的系统信息到指定的地址
- 只有本次获取的系统信息和上次不一样时才会触发上报
- 上报示例：
```bash
{"hostname":"ubuntu-PC","inner_ip":"[2409","cpu_num":8,"mem_total":"15.3GB","disk":"/dev/nvme0n1=476.9GB,/dev/sda=0.0GB","os_type":"linux","os_version":"Deepin 20.8 kernel: 5.15.77-amd64-desktop","uuid":"8B2CC74C-2D08-11B2-A85C-CDAACAE15DD7","virtual":false,"agentVersion":"1.0.0"}
```

<a name="usage"></a>
## 使用方法
在启动node_exporter时加上参数`--cmdb.address`