package model

type Request struct {
	//Image    string `json:"Image,omitempty"`
	Gpu      int    `json:"Gpu,omitempty"`
	Node     string `json:"NodeSelector,omitempty"`
	ShmLimit string `json:"ShmLimit,omitempty"`
}
