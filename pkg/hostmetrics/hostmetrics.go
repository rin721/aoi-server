package hostmetrics

import (
	"context"
	"math"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

const cpuSampleInterval = 200 * time.Millisecond

// Snapshot 是当前主机资源采样快照。
type Snapshot struct {
	CPU  CPUInfo
	Disk []DiskInfo
	RAM  RAMInfo
}

// CPUInfo 描述主机 CPU 核心数和按逻辑核心采样的使用率。
type CPUInfo struct {
	Cores   int
	Percent []float64
}

// RAMInfo 描述主机内存使用情况。
type RAMInfo struct {
	TotalMB     uint64
	UsedMB      uint64
	UsedPercent float64
}

// DiskInfo 描述一个可读挂载点的磁盘使用情况。
type DiskInfo struct {
	FSType      string
	MountPoint  string
	TotalGB     uint64
	TotalMB     uint64
	UsedGB      uint64
	UsedMB      uint64
	UsedPercent float64
}

// Collect 采集当前主机的 CPU、内存和磁盘信息。
func Collect(ctx context.Context) Snapshot {
	return Snapshot{
		CPU:  collectCPU(ctx),
		Disk: collectDisks(),
		RAM:  collectRAM(),
	}
}

func collectCPU(ctx context.Context) CPUInfo {
	cores, err := cpu.Counts(false)
	if err != nil || cores < 1 {
		cores = runtime.NumCPU()
	}

	select {
	case <-ctx.Done():
		return CPUInfo{Cores: cores}
	default:
	}

	percentages, err := cpu.Percent(cpuSampleInterval, true)
	if err != nil {
		return CPUInfo{Cores: cores}
	}
	out := make([]float64, 0, len(percentages))
	for _, value := range percentages {
		out = append(out, roundPercent(value))
	}
	return CPUInfo{
		Cores:   cores,
		Percent: out,
	}
}

func collectRAM() RAMInfo {
	stats, err := mem.VirtualMemory()
	if err != nil || stats == nil {
		return RAMInfo{}
	}
	return RAMInfo{
		TotalMB:     bytesToMB(stats.Total),
		UsedMB:      bytesToMB(stats.Used),
		UsedPercent: roundPercent(stats.UsedPercent),
	}
}

func collectDisks() []DiskInfo {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return []DiskInfo{}
	}

	out := make([]DiskInfo, 0, len(partitions))
	seen := make(map[string]struct{}, len(partitions))
	for _, partition := range partitions {
		mountPoint := strings.TrimSpace(partition.Mountpoint)
		if mountPoint == "" {
			continue
		}
		key := strings.ToLower(mountPoint)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		usage, err := disk.Usage(mountPoint)
		if err != nil || usage == nil || usage.Total == 0 {
			continue
		}
		fsType := strings.TrimSpace(usage.Fstype)
		if fsType == "" {
			fsType = strings.TrimSpace(partition.Fstype)
		}
		out = append(out, DiskInfo{
			FSType:      fsType,
			MountPoint:  mountPoint,
			TotalGB:     bytesToGB(usage.Total),
			TotalMB:     bytesToMB(usage.Total),
			UsedGB:      bytesToGB(usage.Used),
			UsedMB:      bytesToMB(usage.Used),
			UsedPercent: roundPercent(usage.UsedPercent),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].MountPoint < out[j].MountPoint
	})
	return out
}

func bytesToMB(value uint64) uint64 {
	const bytesPerMB = 1024 * 1024
	return value / bytesPerMB
}

func bytesToGB(value uint64) uint64 {
	const bytesPerGB = 1024 * 1024 * 1024
	return value / bytesPerGB
}

func roundPercent(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0
	}
	return math.Round(value*10) / 10
}
