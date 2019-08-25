package config

type Config struct {
	Namespace    string
	Pod          string
	Svc          string
	Image        string
	Port         int
	ExternalPort int
	Timout       int
}
