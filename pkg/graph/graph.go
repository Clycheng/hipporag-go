package graph

// graph.go - 知识图谱
// 用途：构建和管理实体关系图谱，支持多种类型的节点和边
// 主要功能：
// - AddNode: 添加节点（实体或文档块）
// - AddEdge: 添加边（事实关系、段落关系、同义关系）
// - GetNeighbors: 获取节点的邻居
// - 支持并发安全的图操作

import "sync"

// 知识图谱

type Graph struct {
	nodes   map[string]*Node            // 节点：实体 + 文档块
	edges   map[string]map[string]*Edge // 边：事实边 + 段落边 + 同义边
	adjList map[string][]string         // 邻接表（用于 PPR）
	mu      sync.RWMutex
}

type Node struct {
	ID      string
	Content string
	Type    string // "entity" 或 "chunk"
}

type Edge struct {
	From   string
	To     string
	Weight float64
	Type   string // "fact", "passage", "synonymy"
}

// NewGraph 创建新的知识图谱
func NewGraph() *Graph {
	return &Graph{
		nodes:   make(map[string]*Node),
		edges:   make(map[string]map[string]*Edge),
		adjList: make(map[string][]string),
	}
}

// AddNode 添加节点
func (g *Graph) AddNode(id, content, nodeType string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.nodes[id] = &Node{
		ID:      id,
		Content: content,
		Type:    nodeType,
	}

	// 初始化邻接表
	if _, exists := g.adjList[id]; !exists {
		g.adjList[id] = []string{}
	}
}

// AddEdge 添加边
func (g *Graph) AddEdge(from, to string, weight float64, edgeType string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 确保节点存在
	if _, exists := g.nodes[from]; !exists {
		return
	}
	if _, exists := g.nodes[to]; !exists {
		return
	}

	// 添加边
	if g.edges[from] == nil {
		g.edges[from] = make(map[string]*Edge)
	}

	g.edges[from][to] = &Edge{
		From:   from,
		To:     to,
		Weight: weight,
		Type:   edgeType,
	}

	// 更新邻接表
	g.adjList[from] = append(g.adjList[from], to)
}

// GetNode 获取节点
func (g *Graph) GetNode(id string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[id]
	return node, exists
}

// GetNeighbors 获取节点的所有邻居
func (g *Graph) GetNeighbors(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.adjList[id]
}

// GetEdge 获取边
func (g *Graph) GetEdge(from, to string) (*Edge, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if edges, exists := g.edges[from]; exists {
		edge, found := edges[to]
		return edge, found
	}
	return nil, false
}

// NodeCount 返回节点数量
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount 返回边数量
func (g *Graph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}
