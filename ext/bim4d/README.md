# FLYWAVE_bim4d_metadata

## 概述

FLYWAVE_bim4d_metadata 扩展为glTF文件提供了BIM 4D（时间和进度管理）数据支持。该扩展允许将施工进度、工作项和相关元数据与3D模型元素关联起来，从而实现4D施工模拟和进度可视化。

## 功能特性

- 定义施工工作项（Work Items）及其时间属性
- 将工作项与glTF节点（模型元素）关联
- 支持工作项元数据的二进制编码和解码
- 提供工作项进度计算和验证功能
- 支持批量模型级别的工作项管理

## 安装

```bash
go get github.com/flywave/gltf/ext/bim4d
```

## 使用方法

### 创建工作项

```go
props := map[string]interface{}{
    "id":           "foundation-001",
    "name":         "浇筑地基",
    "description":  "浇筑建筑地基混凝土",
    "status":       bim4d.InProgress,
    "generateType": bim4d.Manual,
    "workType":     bim4d.Schedule,
    "startTime":    "2023-01-01T08:00:00Z",
    "endTime":      "2023-01-05T17:00:00Z",
    "startValue":   float32(0),
    "endValue":     float32(100),
    "progressType": bim4d.Percentage,
    "total":        float32(100),
    "metadata": map[string]interface{}{
        "contractor": "ABC建筑公司",
        "budget":     50000,
    },
}

work := bim4d.CreateWorkItem(props)
```

### 将工作项添加到节点

```go
// 创建一个glTF文档
doc := &gltf.Document{
    Nodes: []*gltf.Node{
        {Extensions: make(gltf.Extensions)},
    },
    Extensions: make(gltf.Extensions),
}

// 创建一个工作项
work := &bim4d.WorkItem{
    ID:        "wall-001",
    Name:      "砌筑墙体",
    StartTime: "2023-01-10T08:00:00Z",
    EndTime:   "2023-01-15T17:00:00Z",
}

// 将工作项添加到节点
bim4d.AddWorkItemToNode(doc.Nodes[0], work)
```

### 批量添加工作项到文档

```go
// 创建工作项属性
props := []map[string]interface{}{
    {
        "id":        "foundation-001",
        "name":      "浇筑地基",
        "startTime": "2023-01-01T08:00:00Z",
        "endTime":   "2023-01-05T17:00:00Z",
    },
    {
        "id":        "structure-001",
        "name":      "搭建结构",
        "startTime": "2023-01-06T08:00:00Z",
        "endTime":   "2023-01-15T17:00:00Z",
    },
}

// 批量写入工作项
err := bim4d.WriteBatchModelBim4d(doc, props)
if err != nil {
    log.Fatalf("Failed to write BIM4D data: %v", err)
}
```

### 编码和解码元数据

```go
// 编码元数据到二进制格式
err := bim4d.EncodeBIM4dMetadata(doc)
if err != nil {
    log.Fatalf("Failed to encode metadata: %v", err)
}

// 从二进制格式解码元数据
err = bim4d.DecodeBIM4dMetadata(doc)
if err != nil {
    log.Fatalf("Failed to decode metadata: %v", err)
}
```

## 数据结构

### WorkItem

| 字段 | 类型 | 描述 |
|------|------|------|
| ID | string | 工作项唯一标识符 |
| Name | string | 工作项名称 |
| Description | *string | 工作项描述 |
| Status | WorkStatus | 工作项状态 (pending, in_progress, completed) |
| GenerateType | GenerateType | 生成类型 (auto, manual) |
| WorkType | WorkType | 工作项类型 (schedule, plan) |
| StartTime | string | 开始时间 (ISO 8601格式) |
| EndTime | string | 结束时间 (ISO 8601格式) |
| ScheduleStart | *string | 计划开始时间 |
| ScheduleEnd | *string | 计划结束时间 |
| StartValue | float32 | 起始值 |
| EndValue | float32 | 结束值 |
| ProgressType | ProgressType | 进度类型 (percentage, absolute) |
| Total | float32 | 总量 |
| Metadata | map[string]interface{} | 自定义元数据 |
| MetadataBufferView | *uint32 | 指向二进制元数据的缓冲区视图 |

## 测试

要运行测试：

```bash
go test ./ext/bim4d/... -v
```

## 许可证

MIT