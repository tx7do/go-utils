package math

import (
	"math"
)

// Sign 符号函数（Sign function，简称sgn）是一个逻辑函数，用以判断实数的正负号。为避免和英文读音相似的正弦函数（sine）混淆，它亦称为Signum function。
func Sign[T int | int8 | int16 | int32 | int64 | float32 | float64](x T) T {
	switch {
	case x < 0: // x < 0 : -1
		return -1
	case x > 0: // x > 0 : +1
		return +1
	default: // x == 0 : 0
		return 0
	}
}

// Mean 计算给定数据的平均值
func Mean(num []float64) float64 {
	var count = len(num)
	var sum float64 = 0
	for i := 0; i < count; i++ {
		sum += num[i]
	}
	return sum / float64(count)
}

// Variance 使用平均值计算给定数据的方差
func Variance(mean float64, num []float64) float64 {
	var count = len(num)
	var variance float64 = 0
	for i := 0; i < count; i++ {
		variance += math.Pow(num[i]-mean, 2)
	}
	return variance / float64(count)
}

// StandardDeviation 使用方差计算给定数据的标准偏差
func StandardDeviation(num []float64) float64 {
	var mean = Mean(num)
	var variance = Variance(mean, num)
	return math.Sqrt(variance)
}
