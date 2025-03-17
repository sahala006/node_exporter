package reporter

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getlantern/osversion"
	"github.com/jaypipes/ghw"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/node_exporter/g"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/yumaojun03/dmidecode"
	"io/ioutil"
	"log/slog"
	"net"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"
)

var FailCount = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "cmdb_send_fail_count",
	Help: "fail count of cmdb report.",
})

var failCount = 0

type Disk struct {
	devicePath   string
	serial       string
	deviceNumber string
	size         int64
	model        string
	speed        string
}

func getOsVesion() (os_version string, err error) {
	os_version, err = osversion.GetHumanReadable()
	return
}

func GetLinkSpeed(deviceNumber string) string {
	linkNumber := strings.TrimPrefix(deviceNumber, "ata")
	path := fmt.Sprintf("/sys/class/ata_link/link%s/sata_spd", linkNumber)
	spd, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(spd))
}

func GetDiskInfo(fsRoot string) ([]Disk, error) {
	if fsRoot == "" {
		fsRoot = "/"
	}
	// block, err := ghw.Block(ghw.WithChroot(fsRoot))
	block, err := ghw.Block()
	if err != nil {
		return nil, err
	}

	var info []Disk

	for _, disk := range block.Disks {
		devNum := GetDeviceNumber(disk.Name)
		info = append(info, Disk{
			devicePath:   fmt.Sprintf("/dev/%s", disk.Name),
			serial:       disk.SerialNumber,
			size:         int64(disk.SizeBytes),
			model:        disk.Model,
			deviceNumber: devNum,
			speed:        GetLinkSpeed(devNum),
		})
	}
	return info, nil
}

func getLocalIp(cmdbAddr string) (ip string, err error) {
	host := strings.Split(cmdbAddr, "//")[1]
	conn, err := net.DialTimeout("udp", host, time.Second*5)
	if err != nil {
		return
	}
	ip = strings.Split(conn.LocalAddr().String(), ":")[0]
	defer conn.Close()
	return
}

type System struct {
	Hostname     string `mapstructure:"hostname" json:"hostname"`
	Inner_ip     string `mapstructure:"inner_ip" json:"inner_ip"`
	CpuNum       int    `mapstructure:"cpu_num" json:"cpu_num"`
	MemTotal     string `mapstructure:"mem_total" json:"mem_total"`
	Disk         string `mapstructure:"disk" json:"disk"`
	OsType       string `mapstructure:"os_type" json:"os_type"`
	OsVersion    string `mapstructure:"os_version" json:"os_version"`
	Uuid         string `mapstructure:"uuid" json:"uuid"`
	Virtual      bool   `mapstructure:"virtual" json:"virtual"`
	AgentVersion string `mapstructure:"agentVersion" json:"agentVersion"`
}

func getSystemInfo(cmdbAddr string) (systemInfo System, err error) {
	cpuNum := runtime.NumCPU()

	Hostname, err := os.Hostname()
	if err != nil {
		return
	}

	ip, err := getLocalIp(cmdbAddr)
	if err != nil {
		return
	}

	//meminfo := &MemInfo{}
	//meminfo.Update()
	//memTotal := fmt.Sprintf("%sMB", strconv.Itoa(int(meminfo.Total()/1024/1024)))

	m, _ := mem.VirtualMemory()
	memTotal := fmt.Sprintf("%.1fGB", float64(m.Total)/1024/1024/1024)
	os_type := runtime.GOOS

	os_version, err := getOsVesion()
	if err != nil {
		return
	}
	dmi, err := dmidecode.New()
	if err != nil {
		return
	}
	infos, err := dmi.System()
	if err != nil {
		return
	}
	uuid := infos[0].UUID
	productName := infos[0].ProductName
	var virtual = false
	if find := strings.Contains(strings.ToLower(productName), "virtual"); find {
		virtual = true
	}
	if find := strings.Contains(strings.ToLower(productName), "kvm"); find {
		virtual = true
	}
	//获取磁盘信息
	disks, err := GetDiskInfo("/")
	if err != nil {
		return
	}
	var diskList []string
	for _, disk := range disks {
		kv := fmt.Sprintf("%s=%.1fGB", disk.devicePath, float64(disk.size)/1024/1024/1024)
		diskList = append(diskList, kv)
	}
	diskInfo := strings.Join(diskList, ",")
	systemInfo = System{
		Hostname:     Hostname,
		Inner_ip:     ip,
		CpuNum:       cpuNum,
		MemTotal:     memTotal,
		Disk:         diskInfo,
		OsType:       os_type,
		OsVersion:    os_version,
		Uuid:         uuid,
		Virtual:      virtual,
		AgentVersion: g.Version,
	}
	return
}

func sendToCmdb(systemInfo System, cmdbAddr string) (err error) {
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(time.Now().Add(10 * time.Second))
				return c, nil
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	args, _ := json.Marshal(systemInfo)
	request, err := http.NewRequest("POST", cmdbAddr, bytes.NewBuffer([]byte(args)))
	if err != nil {
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var res map[string]interface{}
	err = json.Unmarshal([]byte(string(body)), &res)
	if err != nil {
		return
	}
	if code, ok := res["code"]; ok {
		if code.(float64) != 0 {
			msg, _ := res["msg"]
			err = errors.New(msg.(string))
			return
		}
	} else {
		resJson, _ := json.Marshal(res)
		err = errors.New(string(resJson))
	}
	return
}

func Report(logger *slog.Logger, cmdbAddr string) {
	var curInfo = System{}
	for {
		systemInfo, err := getSystemInfo(cmdbAddr)
		if strings.Contains(systemInfo.Hostname, "test_") {
			time.Sleep(10 * time.Second)
			continue
		}
		if err != nil {
			logger.Info("collector host info fail", "err", err)
			failCount = 6
			FailCount.Set(float64(failCount))
			time.Sleep(10 * time.Second)
			continue
		}
		// Compare the system information obtained this time with the previous one， if the same then ignore else need report
		if !reflect.DeepEqual(curInfo, systemInfo) {
			failCount = 0
			FailCount.Set(float64(failCount))
			for {
				err := sendToCmdb(systemInfo, cmdbAddr)
				if err != nil {
					logger.Error("send to cmdb fall", "err", err)
					failCount++
					FailCount.Set(float64(failCount))
					if failCount >= 3 {
						return
					} else {
						time.Sleep(5 * time.Minute)
					}

				} else {
					failCount = 0
					FailCount.Set(float64(failCount))
					logger.Info("report success")
					break
				}
			}

		}
		curInfo = systemInfo
		time.Sleep(10 * time.Second)
	}

}
