# pkg/hostmetrics - 主机资源采样

`pkg/hostmetrics` 是可复用的主机资源采样包，当前由 System 模块的服务器状态页面使用。它基于 gopsutil 读取 CPU、内存和磁盘信息，并返回项目自有的轻量快照结构。

## API 分类

- 定位：[CONFIRMED] 公共基础设施 API。
- 稳定边界：`Collect`、`Snapshot`、`CPUInfo`、`RAMInfo`、`DiskInfo`。
- 当前风险：[RISK] 采样结果依赖宿主机权限、挂载点和操作系统能力，调用方应允许空值或部分数据。
- 非目标：[CONFIRMED] 本包不做历史指标存储、告警、推送或权限判断。

## 使用示例

```go
snapshot := hostmetrics.Collect(ctx)

fmt.Println(snapshot.CPU.Cores)
fmt.Println(snapshot.RAM.UsedPercent)
for _, disk := range snapshot.Disk {
    fmt.Println(disk.MountPoint, disk.UsedPercent)
}
```

`Collect` 会尊重传入的 `context.Context`。如果上下文已取消，CPU 使用率可能为空，但核心数仍会尽量返回。
