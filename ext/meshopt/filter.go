package meshopt

import (
	"encoding/binary"
	"errors"
	"math"
)

// decodeOctahedral 解码八面体编码的法线数据
// 输入: 压缩后的字节切片 (每4字节一个编码值)
// 输出: 解码后的浮点法线数据 (每3个float32表示一个法线向量)
func decodeOctahedral(data []byte, stride uint32) ([]byte, error) {
	if len(data)%4 != 0 {
		return nil, errors.New("八面体编码数据长度必须是4的倍数")
	}

	if stride < 8 { // 最小输出步长: 3*float32(4字节) + 1字节填充 = 12字节
		stride = 12
	}

	count := len(data) / 4
	result := make([]byte, int(stride)*count)
	offset := 0

	for i := 0; i < count; i++ {
		// 读取编码值 (2个int16)
		x := int16(binary.LittleEndian.Uint16(data[i*4 : i*4+2]))
		y := int16(binary.LittleEndian.Uint16(data[i*4+2 : i*4+4]))

		// 转换为[-1, 1]范围内的浮点数
		fx := float32(x) / 32767.0
		fy := float32(y) / 32767.0

		// 计算Z分量
		fz := 1.0 - float32(math.Abs(float64(fx))) - float32(math.Abs(float64(fy)))

		// 计算最终法线向量
		var nx, ny, nz float32
		if fz >= 0 {
			nx = fx
			ny = fy
			nz = fz
		} else {
			nx = float32(math.Copysign(float64(1.0-fy), float64(fx)))
			ny = float32(math.Copysign(float64(1.0-fx), float64(fy)))
			nz = fz
		}

		// 归一化
		length := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if length > 0 {
			nx /= length
			ny /= length
			nz /= length
		}

		// 写入结果
		binary.LittleEndian.PutUint32(result[offset:], math.Float32bits(nx))
		binary.LittleEndian.PutUint32(result[offset+4:], math.Float32bits(ny))
		binary.LittleEndian.PutUint32(result[offset+8:], math.Float32bits(nz))

		offset += int(stride)
	}

	return result, nil
}

// decodeQuaternion 解码四元数编码的旋转数据
// 输入: 压缩后的字节切片 (每8字节一个编码值)
// 输出: 解码后的四元数数据 (每4个float32表示一个四元数)
func decodeQuaternion(data []byte) ([]byte, error) {
	if len(data)%8 != 0 {
		return nil, errors.New("四元数编码数据长度必须是8的倍数")
	}

	count := len(data) / 8
	result := make([]byte, 16*count) // 每个四元数占16字节 (4*float32)

	for i := 0; i < count; i++ {
		// 读取编码值 (4个int16)
		x := int16(binary.LittleEndian.Uint16(data[i*8 : i*8+2]))
		y := int16(binary.LittleEndian.Uint16(data[i*8+2 : i*8+4]))
		z := int16(binary.LittleEndian.Uint16(data[i*8+4 : i*8+6]))
		w := int16(binary.LittleEndian.Uint16(data[i*8+6 : i*8+8]))

		// 转换为浮点数 [-1, 1] 范围
		fx := float32(x) / 32767.0
		fy := float32(y) / 32767.0
		fz := float32(z) / 32767.0
		fw := float32(w) / 32767.0

		// 计算实际四元数
		length := float32(math.Sqrt(float64(fx*fx + fy*fy + fz*fz + fw*fw)))
		if length > 0 {
			fx /= length
			fy /= length
			fz /= length
			fw /= length
		}

		// 写入结果
		offset := i * 16
		binary.LittleEndian.PutUint32(result[offset:], math.Float32bits(fx))
		binary.LittleEndian.PutUint32(result[offset+4:], math.Float32bits(fy))
		binary.LittleEndian.PutUint32(result[offset+8:], math.Float32bits(fz))
		binary.LittleEndian.PutUint32(result[offset+12:], math.Float32bits(fw))
	}

	return result, nil
}

// decodeExponential 解码指数编码的浮点数据
// 输入: 压缩后的字节切片 (每4字节一个编码值)
// 输出: 解码后的浮点数据 (每4字节一个float32)
func decodeExponential(data []byte) ([]byte, error) {
	if len(data)%4 != 0 {
		return nil, errors.New("指数编码数据长度必须是4的倍数")
	}

	count := len(data) / 4
	result := make([]byte, 4*count)

	for i := 0; i < count; i++ {
		// 读取32位整数
		encoded := binary.LittleEndian.Uint32(data[i*4 : i*4+4])

		// 提取符号位和指数
		sign := encoded >> 31
		exponent := int(encoded>>23) & 0xFF

		// 提取尾数 (23位)
		mantissa := float32(encoded & 0x7FFFFF)

		// 计算浮点值
		var value float32
		switch exponent {
		case 0:
			// 零或次正规数
			value = mantissa * 1.4e-45 // ~2^-149
		case 255:
			// 无穷大或NaN
			value = float32(math.NaN())
		default:
			// 正规数
			mantissa = (1.0 + mantissa/8388608.0) // 2^23
			exponent -= 127
			value = mantissa * float32(math.Pow(2, float64(exponent)))
		}

		// 应用符号
		if sign == 1 {
			value = -value
		}

		// 写入结果
		binary.LittleEndian.PutUint32(result[i*4:], math.Float32bits(value))
	}

	return result, nil
}

// encodeOctahedral 将法线向量编码为八面体表示
// 输入: 原始法线数据 (每3个float32表示一个法线)
// 输出: 编码后的数据 (每4字节一个编码值)
func encodeOctahedral(data []byte, stride uint32) ([]byte, error) {
	if len(data) < 12 {
		return nil, errors.New("数据长度不足")
	}

	// 计算法线数量
	count := len(data) / int(stride)
	result := make([]byte, count*4) // 每个法线编码为4字节

	for i := 0; i < count; i++ {
		offset := i * int(stride)

		// 读取法线分量
		x := math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
		y := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+4:]))
		z := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+8:]))

		// 归一化 (确保是单位向量)
		length := float32(math.Sqrt(float64(x*x + y*y + z*z)))
		if length > 0 {
			x /= length
			y /= length
			z /= length
		}

		// 八面体编码
		invL1 := 1.0 / (abs(x) + abs(y) + abs(z))
		nx := x * invL1
		ny := y * invL1

		if z < 0 {
			tmp := nx
			nx = (1 - abs(ny)) * float32(sign(tmp))
			ny = (1 - abs(tmp)) * float32(sign(ny))
		}

		// 映射到[-1,1]范围
		nx = clamp(nx*0.5+0.5, 0, 1)*2 - 1
		ny = clamp(ny*0.5+0.5, 0, 1)*2 - 1

		// 转换为int16
		ix := int16(nx * 32767)
		iy := int16(ny * 32767)

		// 写入结果
		binary.LittleEndian.PutUint16(result[i*4:], uint16(ix))
		binary.LittleEndian.PutUint16(result[i*4+2:], uint16(iy))
	}

	return result, nil
}

// encodeQuaternion 将四元数编码为紧凑表示
// 输入: 原始四元数数据 (每4个float32表示一个四元数)
// 输出: 编码后的数据 (每8字节一个编码值)
func encodeQuaternion(data []byte) ([]byte, error) {
	if len(data)%16 != 0 {
		return nil, errors.New("四元数数据长度必须是16的倍数")
	}

	count := len(data) / 16
	result := make([]byte, count*8) // 每个四元数编码为8字节

	for i := 0; i < count; i++ {
		offset := i * 16

		// 读取四元数分量
		x := math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
		y := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+4:]))
		z := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+8:]))
		w := math.Float32frombits(binary.LittleEndian.Uint32(data[offset+12:]))

		// 归一化
		length := float32(math.Sqrt(float64(x*x + y*y + z*z + w*w)))
		if length > 0 {
			x /= length
			y /= length
			z /= length
			w /= length
		}

		// 找到最大绝对值的分量
		maxIndex := 0
		maxValue := abs(x)
		if abs(y) > maxValue {
			maxIndex = 1
			maxValue = abs(y)
		}
		if abs(z) > maxValue {
			maxIndex = 2
			maxValue = abs(z)
		}
		if abs(w) > maxValue {
			maxIndex = 3
			maxValue = abs(w)
		}

		// 编码为3个int16 (存储3个分量和最大分量索引)
		var components [3]float32
		switch maxIndex {
		case 0:
			components = [3]float32{y, z, w}
		case 1:
			components = [3]float32{x, z, w}
		case 2:
			components = [3]float32{x, y, w}
		case 3:
			components = [3]float32{x, y, z}
		}

		// 写入结果 (3个分量 + 1字节的索引信息)
		idxByte := byte(maxIndex) | (byte(sign(components[0])) << 2) | (byte(sign(components[1])) << 3) | (byte(sign(components[2])) << 4)
		result[i*8] = idxByte

		// 编码三个分量
		for j := 0; j < 3; j++ {
			val := int16(components[j] * 32767)
			binary.LittleEndian.PutUint16(result[i*8+2+j*2:], uint16(val))
		}
	}

	return result, nil
}

// encodeExponential 将浮点数编码为指数表示
// 输入: 原始浮点数据 (每4字节一个float32)
// 输出: 编码后的数据 (每4字节一个编码值)
func encodeExponential(data []byte) ([]byte, error) {
	if len(data)%4 != 0 {
		return nil, errors.New("浮点数据长度必须是4的倍数")
	}

	count := len(data) / 4
	result := make([]byte, count*4)

	for i := 0; i < count; i++ {
		// 读取原始浮点值
		val := math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))

		// 处理特殊值
		var encoded uint32
		if math.IsNaN(float64(val)) {
			encoded = 0x7FFFFFFF // NaN 表示
		} else if math.IsInf(float64(val), 0) {
			encoded = 0x7F800000 // 无穷大
			if val < 0 {
				encoded |= 0x80000000 // 负号
			}
		} else {
			// 提取符号、指数和尾数
			bits := math.Float32bits(val)
			sign := bits >> 31
			exponent := (bits >> 23) & 0xFF
			mantissa := bits & 0x7FFFFF

			// 指数编码
			if exponent == 0 && mantissa != 0 {
				// 次正规数 - 转换为正规数
				exponent = 1
				for mantissa < 0x400000 {
					mantissa <<= 1
					exponent--
				}
				mantissa &= 0x7FFFFF
			}

			// 组合编码值
			encoded = (sign << 31) | (exponent << 23) | mantissa
		}

		// 写入结果
		binary.LittleEndian.PutUint32(result[i*4:], encoded)
	}

	return result, nil
}

// 辅助函数
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x float32) int32 {
	if x < 0 {
		return -1
	}
	return 1
}

func clamp(x, min, max float32) float32 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
