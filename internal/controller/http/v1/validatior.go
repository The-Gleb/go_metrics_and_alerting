package v1

type Validator interface {
	Valid() (problems map[string]string)
}
