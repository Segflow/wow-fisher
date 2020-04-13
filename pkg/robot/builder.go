package robot

var (
	builders = make(map[string]ActionBuilder)
)

type ActionBuilder interface {
	Build(map[string]string) (Action, error)
}

func builderFor(name string) ActionBuilder {
	return builders[name]
}

type builderFunc func(map[string]string) (Action, error)

func (f builderFunc) Build(params map[string]string) (Action, error) {
	return f(params)
}

func RegisterBuilderFunc(name string, f builderFunc) {
	builders[name] = f
}
