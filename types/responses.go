package types

type ReqMessage struct {
	Code    uint64 `json:"code" comment:"返回码"`
	Message string `json:"message" comment:"返回信息"`
	Data    string `json:"data" comment:"返回数据"`
}
