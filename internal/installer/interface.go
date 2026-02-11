package installer

type Installer interface {
	Download() (string, error)
	Extract(src string) error
}
