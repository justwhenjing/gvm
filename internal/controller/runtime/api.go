package runtime

type IRuntime interface {
	List(filter string) error
	Use(version string) error
	Install(version string) error
	Uninstall(version string) error
}
