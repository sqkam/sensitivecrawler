package result

type Result struct {
	// 爬取站点
	Site string `json:"site,omitempty"`
	// url
	Url string `json:"url,omitempty"`

	// 敏感信息
	Info       string      `json:"info,omitempty"`
	Statistics *Statistics `json:"statistics,omitempty"`
}

type Statistics struct {
	UrlCount int64 `json:"url_count,omitempty"`
	// 内存用量
	MemoryTotal int64 `json:"memory_total,omitempty"`
}
