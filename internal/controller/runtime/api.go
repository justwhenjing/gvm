package runtime

type IRuntime interface {
	ListVersions(filter string) error
}
