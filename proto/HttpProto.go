package proto

// @Description:接口基础返回参数
type HttpBasicResp struct {
	Ec      int    `json:"ec"`
	Em      string `json:"em"`
	Timesec int    `json:"timesec"`
}
