# KHR_gaussian_splatting 扩展

## 概述

KHR_gaussian_splatting 扩展为 glTF 格式提供了高斯泼溅（Gaussian Splatting）的支持。高斯泼溅是一种用于 3D 点云渲染的技术，可以生成高质量的 3D 场景表示。

## 规范

该扩展实现了 [KHR_gaussian_splatting](https://github.com/KhronosGroup/glTF/tree/main/extensions/2.0/Vendor/KHR_gaussian_splatting) 草案规范。

## 数据映射

| 泼溅数据 | glTF 属性 |
|---------|----------|
| 位置 (Position) | POSITION |
| 颜色 (Color) | COLOR_0 RGB 通道 |
| 不透明度 (Opacity) | COLOR_0 A 通道 |
| 旋转 (Rotation) | _ROTATION |
| 缩放 (Scale) | _SCALE |

## 使用示例

### 创建高斯泼溅模型

```go
import "github.com/flywave/gltf/ext/splatting"

// 创建顶点数据
vertexData := &splatting.VertexData{
    Positions: []float32{0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0, 0.0},
    Colors:    []float32{1.0, 0.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0, 1.0},
    Scales:    []float32{0.1, 0.1, 0.1, 0.2, 0.2, 0.2, 0.3, 0.3, 0.3},
    Rotations: []float32{1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0},
}

// 创建 glTF 文档
doc := &gltf.Document{
    Asset: gltf.Asset{Version: "2.0"},
}

// 连接高斯泼溅数据
splatting.WireGaussianSplatting(doc, vertexData, false)
```

### 读取高斯泼溅数据

```go
import "github.com/flywave/gltf/ext/splatting"

// 打开 glTF 文件
doc, err := gltf.Open("model.gltf")
if err != nil {
    // 处理错误
}

// 获取图元
primitive := doc.Meshes[0].Primitives[0]

// 读取高斯泼溅数据
data, err := splatting.ReadGaussianSplatting(doc, primitive)
if err != nil {
    // 处理错误
}

// 使用数据
vertexCount := len(data.Positions) / 3
```

## API 参考

### 类型定义

#### VertexData
```go
type VertexData struct {
    Positions []float32  // 顶点位置 (XYZ)
    Colors    []float32  // 顶点颜色 (RGBA)
    Scales    []float32  // 顶点缩放 (XYZ)
    Rotations []float32  // 顶点旋转 (四元数 XYZW)
}
```

#### GaussianSplatting
```go
type GaussianSplatting struct{}
```

### 函数

#### WireGaussianSplatting
```go
func WireGaussianSplatting(
    doc *gltf.Document,
    vertexData *VertexData,
    compress bool
) (*GaussianSplatting, error)
```
将顶点数据连接到 glTF 文档中。

参数:
- `doc`: glTF 文档
- `vertexData`: 顶点数据
- `compress`: 是否使用压缩

返回:
- `*GaussianSplatting`: 高斯泼溅扩展实例
- `error`: 错误信息

#### ReadGaussianSplatting
```go
func ReadGaussianSplatting(
    doc *gltf.Document,
    primitive *gltf.Primitive
) (*VertexData, error)
```
从 glTF 文档中读取高斯泼溅数据。

参数:
- `doc`: glTF 文档
- `primitive`: 图元

返回:
- `*VertexData`: 顶点数据
- `error`: 错误信息

#### ValidateRotation
```go
func ValidateRotation(rotations []float32) error
```
验证旋转数据是否为单位四元数。

参数:
- `rotations`: 旋转数据（四元数）

返回:
- `error`: 错误信息

## 测试

运行单元测试:
```bash
go test ./ext/splatting/... -v
```

运行集成测试:
```bash
go test ./ext/splatting/... -v -run Test.*Integration
```

运行示例测试:
```bash
go test ./ext/splatting/... -v -run Example
```

## 已知实现

该扩展目前在以下项目中实现：
- [CesiumJS](https://github.com/CesiumGS/cesium/tree/splat-shader)

## 资源

- [GaussianSplats3D](https://github.com/mkkellogg/GaussianSplats3D/issues/47#issuecomment-1801360116)