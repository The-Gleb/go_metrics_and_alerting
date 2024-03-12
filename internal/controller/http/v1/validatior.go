package http

type Validator interface {
	Valid() (problems map[string]string)
}
