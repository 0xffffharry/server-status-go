package core

import (
	"context"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/anatol/smart.go"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

func GetCPUUsage(ctx context.Context) (float64, error) { // %
	percents, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return 0, err
	}
	allPercent := 0.0
	for _, percent := range percents {
		allPercent += percent
	}
	return allPercent / float64(len(percents)), nil
}

func GetCPULoad(ctx context.Context) (float64, float64, float64, error) { // 1min/5min/15min
	info, err := load.Avg()
	if err != nil {
		return 0, 0, 0, err
	}
	return info.Load1, info.Load5, info.Load15, nil
}

type RAMInfo struct {
	Total uint64 // (B)
	Used  uint64 // (B)
	Free  uint64 // (B)

	SwapTotal uint64 // (B)
	SwapFree  uint64 // (B)
}

func GetRAMUsage(ctx context.Context) (*RAMInfo, error) { // total/used/available (B)
	info, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return &RAMInfo{
		Total:     info.Total,
		Used:      info.Used,
		Free:      info.Free,
		SwapTotal: info.SwapTotal,
		SwapFree:  info.SwapFree,
	}, nil
}

type NetworkInfo struct {
	Name      string `json:"name"`
	BytesSent uint64 `json:"sent"` // (B)
	BytesRecv uint64 `json:"recv"` // (B)
}

func GetNetInfo(ctx context.Context) ([]NetworkInfo, error) {
	infos, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
	}
	var ret []NetworkInfo
	for _, info := range infos {
		ret = append(ret, NetworkInfo{
			Name:      info.Name,
			BytesSent: info.BytesSent,
			BytesRecv: info.BytesRecv,
		})
	}
	return ret, nil
}

type SystemInfo struct {
	Hostname      string        `json:"hostname"`
	Uptime        time.Duration `json:"uptime"` // second
	OS            string        `json:"os"`
	OSVersion     string        `json:"os_version"`
	KernelVersion string        `json:"kernel_version"`
	Arch          string        `json:"arch"`
}

func GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}
	os := info.OS
	if info.Platform != "" {
		os += " - " + info.Platform
	}
	return &SystemInfo{
		Hostname:      info.Hostname,
		Uptime:        time.Duration(info.Uptime),
		OS:            os,
		OSVersion:     info.PlatformVersion,
		KernelVersion: info.KernelVersion,
		Arch:          info.KernelArch,
	}, nil
}

var (
	diskCache     []DiskInfo
	diskCacheN    uint64
	diskCacheMaxN uint64 = 60
)

type DiskInfo struct {
	Device string `json:"device"`
	Path   string `json:"path"`
	Total  uint64 `json:"total"` // (B)
	Used   uint64 `json:"used"`  // (B)
}

func GetDisk(ctx context.Context) ([]DiskInfo, error) {
	if len(diskCache) > 0 {
		diskCacheN++
		if diskCacheN < diskCacheMaxN {
			return diskCache, nil
		}
		diskCacheN = 0
	}
	parts, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool)
	var ret []DiskInfo
	for _, part := range parts {
		if m[part.Mountpoint] {
			continue
		}
		m[part.Mountpoint] = true
		d := DiskInfo{
			Device: part.Device,
			Path:   part.Mountpoint,
		}
		usage, err := disk.UsageWithContext(ctx, part.Mountpoint)
		if err == nil {
			d.Total = usage.Total
			d.Used = usage.Used
		}
		ret = append(ret, d)
	}
	diskCache = ret
	return ret, nil
}

var (
	sdNumRegex   = regexp.MustCompile(`\d+$`)
	nvmeNumRegex = regexp.MustCompile(`p\d+$`)

	diskTemperatureCache []TemperatureInfo
	diskTemperatureN     uint64
	diskTemperatureMaxN  uint64 = 60
)

type TemperatureInfo struct {
	Key         string  `json:"key"`
	Temperature float64 `json:"temperature"` // °C
}

func GetTemperature(ctx context.Context) ([]TemperatureInfo, error) {
	infos, err := host.SensorsTemperaturesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	var ret []TemperatureInfo
	for _, info := range infos {
		ret = append(ret, TemperatureInfo{
			Key:         info.SensorKey,
			Temperature: info.Temperature,
		})
	}
	// Disk
	func() {
		if len(diskTemperatureCache) > 0 {
			diskTemperatureN++
			if diskTemperatureN < diskTemperatureMaxN {
				ret = append(ret, diskTemperatureCache...)
				return
			}
			diskTemperatureN = 0
		}
		var diskTemperature []TemperatureInfo
		filepath.WalkDir("/dev", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			var ok bool
			switch {
			case strings.HasPrefix(path, "/dev/sd") && sdNumRegex.MatchString(path):
				ok = true
			case strings.HasPrefix(path, "/dev/nvme") && nvmeNumRegex.MatchString(path):
				ok = true
			}
			if ok {
				dev, err := smart.Open(path)
				if err == nil {
					attrs, err := dev.ReadGenericAttributes()
					if err == nil {
						// Temperature
						diskTemperature = append(diskTemperature, TemperatureInfo{
							Key:         path,
							Temperature: float64(attrs.Temperature),
						})
					}
				}
			}
			return nil
		})
		diskTemperatureCache = diskTemperature
		ret = append(ret, diskTemperature...)
	}()

	return ret, nil
}

func GetAll(ctx context.Context) map[string]any {
	m := make(map[string]any)
	// System
	systemInfo, err := GetSystemInfo(ctx)
	if err == nil {
		m["system"] = systemInfo
	} else {
		m["system"] = nil
	}

	// CPU (%)
	cpuUsage, err := GetCPUUsage(ctx)
	if err == nil {
		m["cpu_usage"] = cpuUsage
	} else {
		m["cpu_usage"] = nil
	}

	cpuLoad1min, cpuLoad5min, cpuLoad15min, err := GetCPULoad(ctx)
	if err == nil {
		m["cpu_load_1min"] = cpuLoad1min
		m["cpu_load_5min"] = cpuLoad5min
		m["cpu_load_15min"] = cpuLoad15min
	} else {
		m["cpu_load_1min"] = nil
		m["cpu_load_5min"] = nil
		m["cpu_load_15min"] = nil
	}

	// RAM (B)
	ramInfo, err := GetRAMUsage(ctx)
	if err == nil {
		m["ram_total"] = ramInfo.Total
		m["ram_used"] = ramInfo.Used
		m["ram_free"] = ramInfo.Free
		m["swap_total"] = ramInfo.SwapTotal
		m["swap_free"] = ramInfo.SwapFree
	} else {
		m["ram_total"] = nil
		m["ram_used"] = nil
		m["ram_free"] = nil
		m["swap_total"] = nil
		m["swap_free"] = nil
	}

	// Network
	netInfos, err := GetNetInfo(ctx)
	if err == nil {
		m["network"] = netInfos
	} else {
		m["network"] = nil
	}

	// Disk
	diskInfos, err := GetDisk(ctx)
	if err == nil {
		m["disk"] = diskInfos
	} else {
		m["disk"] = nil
	}

	// Temperature (°C)
	temperatureInfos, err := GetTemperature(ctx)
	if err == nil {
		m["temperature"] = temperatureInfos
	} else {
		m["temperature"] = nil
	}

	return m
}
