package result

type Result struct {
	// 爬取站点
	Site string
	// url
	Url string
	// 内存用量
	MemoryTotal int64
	// 敏感信息
	Info string
}
