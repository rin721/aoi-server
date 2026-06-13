package processx

import "github.com/shirou/gopsutil/v3/process"

// CreateTime 返回指定 PID 的进程创建时间，单位沿用 gopsutil 的毫秒时间戳。
func CreateTime(pid int) (int64, error) {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return 0, err
	}
	return proc.CreateTime()
}

// IsRunning 校验指定 PID 是否仍在运行；createTime 大于 0 时会同时比对创建时间，避免 PID 复用误判。
func IsRunning(pid int, createTime int64) (bool, error) {
	if pid <= 0 {
		return false, nil
	}
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return false, nil
	}
	running, err := proc.IsRunning()
	if err != nil || !running {
		return false, err
	}
	if createTime <= 0 {
		return true, nil
	}
	actualCreateTime, err := proc.CreateTime()
	if err != nil {
		return false, err
	}
	return actualCreateTime == createTime, nil
}
