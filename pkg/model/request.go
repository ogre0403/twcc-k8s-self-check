package model

type Request struct {
	Gpu      string `json:"gpu,omitempty"`
	Node     string `json:"node,omitempty"`
	ShmLimit string `json:"shm,omitempty"`
}
