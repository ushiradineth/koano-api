package util

type Error struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Error interface{} `json:"error"`
}

type Response struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}
