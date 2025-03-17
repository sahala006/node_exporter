[//]: # (<div align="left"><img src="docs/images/funasr_logo.jpg" width="400"/></div>)

(English|[简体中文](./README_zh.md))

<a name="introduction"></a>
## Function Introduction
Based on the [node_exporter v1.9.0](https://github.com/prometheus/node_exporter/tree/v1.9.0 "node_exporter v1.9.0") for secondary development
- report the system information of the host to the target address regularly
- The reporting will only be triggered when the system information obtained this time is different from the last time
- report example：
```bash
{"hostname":"ubuntu-PC","inner_ip":"[2409","cpu_num":8,"mem_total":"15.3GB","disk":"/dev/nvme0n1=476.9GB,/dev/sda=0.0GB","os_type":"linux","os_version":"Deepin 20.8 kernel: 5.15.77-amd64-desktop","uuid":"8B2CC74C-2D08-11B2-A85C-CDAACAE15DD7","virtual":false,"agentVersion":"1.0.0"}
```

<a name="usage"></a>
## Usage
Add the parameter `--cmdb.address` when start node_exporter