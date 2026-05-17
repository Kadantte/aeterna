package common

type Module[T any] interface {
	Name() string
	Section() string
	LoadAndValidate() (T, error)
}

type RuntimeModule interface {
	Name() string
	Section() string
	LoadAndValidateAny() (any, error)
}

type runtimeModule[T any] struct {
	module Module[T]
}

var modules []RuntimeModule

func Register[T any](module Module[T]) {
	modules = append(modules, runtimeModule[T]{module: module})
}

func RegisteredModules() []RuntimeModule {
	out := make([]RuntimeModule, len(modules))
	copy(out, modules)
	return out
}

func (r runtimeModule[T]) Name() string {
	return r.module.Name()
}

func (r runtimeModule[T]) Section() string {
	return r.module.Section()
}

func (r runtimeModule[T]) LoadAndValidateAny() (any, error) {
	section, err := r.module.LoadAndValidate()
	if err != nil {
		return nil, err
	}
	return section, nil
}
