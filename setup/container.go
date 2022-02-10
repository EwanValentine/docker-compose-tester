package setup

// Publisher -
type Publisher struct {
	URL           string
	TargetPort    int
	PublishedPort int
	Protocol      string
}

// Container is a running container, which is returned by running
// `docker compose ps --format=json`, which is then parsed into
// this struct
type Container struct {
	ID         string
	Name       string
	Command    string
	Project    string
	Service    string
	State      string // @todo - could be an enum
	Health     string
	ExitCode   int
	Publishers []Publisher
}
