package rand

// AliasTable 用于实现基于别名方法的加权随机选择算法，提供 O(1) 时间复杂度的随机选择
type AliasTable struct {
	prob    []float32 // 每个选项的概率
	alias   []int     // 矩形中选项的索引
	origIdx []int     // 映射回原始数组的真实索引
}

// NewAliasTable 创建一个新的别名表，接受一个权重数组，返回一个构建好的别名表。它会过滤掉无效权重（小于等于0的项），并保留原始索引映射，以确保随机选择时能够正确返回原始索引。支持整数权重、float32权重和float64权重三种类型。
func (r *Randomizer) NewAliasTable(weights []int) *AliasTable {
	// 过滤无效权重，保留真实索引
	validWeights, validIndices := filterValidWeights(weights)
	if len(validWeights) == 0 {
		return nil
	}

	scaled := r.aliasCalcScaledWeights(validWeights)
	small, large := r.aliasSplitSmallLarge(scaled)
	prob, alias := r.aliasBuildTable(len(validWeights), scaled, small, large)

	return &AliasTable{
		prob:    prob,
		alias:   alias,
		origIdx: validIndices,
	}
}

// NewAliasTableFloat32 创建一个新的别名表，接受一个float32类型的权重数组，返回一个构建好的别名表。它会过滤掉无效权重（小于等于0.0000001的项），并保留原始索引映射，以确保随机选择时能够正确返回原始索引。
func (r *Randomizer) NewAliasTableFloat32(weights []float32) *AliasTable {
	validWeights, validIndices := filterValidWeightsFloat32(weights)
	if len(validWeights) == 0 {
		return nil
	}

	scaled := r.aliasCalcScaledWeightsFloat32(validWeights)
	small, large := r.aliasSplitSmallLarge(scaled)
	prob, alias := r.aliasBuildTable(len(validWeights), scaled, small, large)

	return &AliasTable{
		prob:    prob,
		alias:   alias,
		origIdx: validIndices,
	}
}

// NewAliasTableFloat64 创建一个新的别名表，接受一个float64类型的权重数组，返回一个构建好的别名表。它会过滤掉无效权重（小于等于0.0000001的项），并保留原始索引映射，以确保随机选择时能够正确返回原始索引。
func (r *Randomizer) NewAliasTableFloat64(weights []float64) *AliasTable {
	validWeights, validIndices := filterValidWeightsFloat64(weights)
	if len(validWeights) == 0 {
		return nil
	}

	scaled := r.aliasCalcScaledWeightsFloat64(validWeights)
	small, large := r.aliasSplitSmallLarge(scaled)
	prob, alias := r.aliasBuildTable(len(validWeights), scaled, small, large)

	return &AliasTable{
		prob:    prob,
		alias:   alias,
		origIdx: validIndices,
	}
}

// AliasChoice 使用别名表进行随机选择，返回原始索引，支持空表和无效表的安全处理
func (r *Randomizer) AliasChoice(at *AliasTable) int {
	if at == nil || len(at.prob) == 0 || len(at.origIdx) == 0 {
		return -1
	}

	n := len(at.prob)
	idx := r.IntN(n)
	if r.Float32() < at.prob[idx] {
		return at.origIdx[idx]
	}
	return at.origIdx[at.alias[idx]]
}

// aliasCalcScaledWeights 计算缩放权重，保留原始索引映射
func (r *Randomizer) aliasCalcScaledWeights(weights []int) []float32 {
	n := len(weights)
	sum := 0
	for _, w := range weights {
		sum += w
	}
	avg := float32(sum) / float32(n)
	scaled := make([]float32, n)
	for i, w := range weights {
		scaled[i] = float32(w) / avg
	}
	return scaled
}

// aliasCalcScaledWeightsFloat32 计算缩放权重，保留原始索引映射
func (r *Randomizer) aliasCalcScaledWeightsFloat32(weights []float32) []float32 {
	n := len(weights)
	sum := float32(0)
	for _, w := range weights {
		sum += w
	}
	avg := sum / float32(n)
	scaled := make([]float32, n)
	for i, w := range weights {
		scaled[i] = w / avg
	}
	return scaled
}

// aliasCalcScaledWeightsFloat64 计算缩放权重，保留原始索引映射
func (r *Randomizer) aliasCalcScaledWeightsFloat64(weights []float64) []float32 {
	n := len(weights)
	sum := float64(0)
	for _, w := range weights {
		sum += w
	}
	avg := sum / float64(n)
	scaled := make([]float32, n)
	for i, w := range weights {
		scaled[i] = float32(w / avg)
	}
	return scaled
}

// aliasSplitSmallLarge 将缩放后的权重分为小于1和大于等于1两类，保留原始索引映射
func (r *Randomizer) aliasSplitSmallLarge(scaled []float32) (small, large []int) {
	for i, sw := range scaled {
		if sw < 1.0 {
			small = append(small, i)
		} else {
			large = append(large, i)
		}
	}
	return
}

// aliasBuildTable 构建别名表，返回概率和别名数组，保留原始索引映射
func (r *Randomizer) aliasBuildTable(n int, scaled []float32, small, large []int) ([]float32, []int) {
	prob := make([]float32, n)
	alias := make([]int, n)

	for len(small) > 0 && len(large) > 0 {
		s := small[len(small)-1]
		small = small[:len(small)-1]
		l := large[len(large)-1]
		large = large[:len(large)-1]

		prob[s] = scaled[s]
		alias[s] = l

		scaled[l] = (scaled[l] + scaled[s]) - 1.0
		if scaled[l] < 1.0 {
			small = append(small, l)
		} else {
			large = append(large, l)
		}
	}

	for _, idx := range large {
		prob[idx] = 1.0
	}
	for _, idx := range small {
		prob[idx] = 1.0
	}

	return prob, alias
}

// filterValidWeights 过滤掉权重小于等于0的项，保留原始索引映射
func filterValidWeights(weights []int) ([]int, []int) {
	var w []int
	var idx []int
	for i, v := range weights {
		if v > 0 {
			w = append(w, v)
			idx = append(idx, i)
		}
	}
	return w, idx
}

// filterValidWeightsFloat32 过滤掉权重小于等于0.0000001的项，保留原始索引映射
func filterValidWeightsFloat32(weights []float32) ([]float32, []int) {
	var w []float32
	var idx []int
	for i, v := range weights {
		if v > 0.0000001 {
			w = append(w, v)
			idx = append(idx, i)
		}
	}
	return w, idx
}

// filterValidWeightsFloat64 过滤掉权重小于等于0.0000001的项，保留原始索引映射
func filterValidWeightsFloat64(weights []float64) ([]float64, []int) {
	var w []float64
	var idx []int
	for i, v := range weights {
		if v > 0.0000001 {
			w = append(w, v)
			idx = append(idx, i)
		}
	}
	return w, idx
}
