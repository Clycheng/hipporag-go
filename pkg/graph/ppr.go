package graph

// ppr.go - Personalized PageRank 算法
// 用途：在知识图谱上执行个性化 PageRank，用于图检索
// 主要功能：
// - PPR: 从种子节点出发，计算所有节点的重要性分数
// - 支持自定义阻尼系数、迭代次数和收敛阈值

import "math"

// PPR 执行 Personalized PageRank 算法
// seedWeights: 种子节点及其初始权重 (例如：查询相关的实体)
// damping: 阻尼系数，通常为 0.5-0.85，控制随机游走的跳转概率
// maxIter: 最大迭代次数
// tolerance: 收敛阈值，当分数变化小于此值时停止迭代
// 返回：所有节点的 PageRank 分数
func (g *Graph) PPR(
	seedWeights map[string]float64,
	damping float64,
	maxIter int,
	tolerance float64,
) map[string]float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(seedWeights) == 0 {
		return make(map[string]float64)
	}

	// 初始化分数
	scores := make(map[string]float64)
	newScores := make(map[string]float64)

	// 归一化种子权重
	totalSeedWeight := 0.0
	for _, weight := range seedWeights {
		totalSeedWeight += weight
	}

	normalizedSeeds := make(map[string]float64)
	for id, weight := range seedWeights {
		normalizedSeeds[id] = weight / totalSeedWeight
		scores[id] = normalizedSeeds[id]
	}

	// 迭代计算 PageRank
	for iter := 0; iter < maxIter; iter++ {
		// 重置新分数
		for id := range scores {
			newScores[id] = 0
		}

		// 对每个节点，将其分数分配给邻居
		for nodeID, score := range scores {
			neighbors := g.adjList[nodeID]

			if len(neighbors) == 0 {
				// 没有出边，分数回流到种子节点
				for seedID, seedWeight := range normalizedSeeds {
					newScores[seedID] += score * seedWeight
				}
			} else {
				// 平均分配给邻居（可以根据边权重加权）
				sharePerNeighbor := score / float64(len(neighbors))
				for _, neighborID := range neighbors {
					newScores[neighborID] += sharePerNeighbor
				}
			}
		}

		// 应用阻尼和种子节点
		// newScore = (1-damping) * seedScore + damping * propagatedScore
		for id := range newScores {
			seedScore := normalizedSeeds[id]
			newScores[id] = (1-damping)*seedScore + damping*newScores[id]
		}

		// 检查收敛
		converged := true
		maxDiff := 0.0
		for id, newScore := range newScores {
			diff := math.Abs(newScore - scores[id])
			if diff > maxDiff {
				maxDiff = diff
			}
			if diff > tolerance {
				converged = false
			}
		}

		// 更新分数
		scores, newScores = newScores, scores

		if converged {
			break
		}
	}

	return scores
}
