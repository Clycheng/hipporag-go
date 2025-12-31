package data

// TestDocuments 测试文档集合
// 用于对比传统 RAG 和 HippoRAG 的检索效果
var TestDocuments = []string{
	"阿尔伯特·爱因斯坦是著名的物理学家",
	"爱因斯坦于1879年3月14日出生于德国乌尔姆",
	"他在1905年发表了相对论",
	"19世纪是指1801年到1900年这段时期",
	"爱因斯坦在1921年获得诺贝尔物理学奖",
}

// TestQuestion 测试问题
var TestQuestion = "爱因斯坦出生于哪个世纪？"

// ExpectedAnswer 期望答案
// 需要结合文档2（出生年份1879）和文档4（19世纪定义）才能回答
var ExpectedAnswer = "爱因斯坦出生于19世纪（1879年）"
