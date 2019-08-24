package model

type CheckResult struct {
	NamespaceCreate    string `json:"CreateNamespace,omitempty"`
	PodCreate          string `json:"CreatePod,omitempty"`
	SvcCreate          string `json:"CreateSVC,omitempty"`
	IntranetConnection string `json:"IntraConnection,omitempty"`
	InternetConnection string `json:"InterConnection,omitempty"`
	ErrorMsg           string `json:"ErrorMessage,omitempty"`
}
