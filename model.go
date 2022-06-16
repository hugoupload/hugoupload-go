package hugoPartUpload

type data struct {
	UploadId int `json:"uploadId"`
}

type response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type initData struct {
	response
	Data data `json:"data"`
}

type completeData struct {
	response
	Data []int `json:"data"`
}

// 默认结构体配置文件
type PartClient struct {
	Token       string // 必填
	Identifier  string // 必填
	User        string // 必填
	Title       string
	Audio       string
	Rule        string   // 必填
	Cat         string   // 必填
	Subcat      []string // 必填
	Actor       string
	Domain      string
	Filename    string // 必填
	Cover       string // 封面图
	UploadId    int
	NewFilename string
	tp          int
	UnAudit     int // 是否需要审核   0: 需要审核  1: 不需要审核
	FtpUserId   int    // ftpuserid
}
