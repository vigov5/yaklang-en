package ssaapi

type OperationConfig struct {
	// limits the recursion depth. Every time the core function is recursed, the counter will be incremented by one.
	// The context counter is subject to this restriction.
	MaxDepth int
}

type OperationOption func(*OperationConfig)

func WithMaxDepth(maxDepth int) OperationOption {
	return func(operationConfig *OperationConfig) {
		operationConfig.MaxDepth = maxDepth
	}
}

func NewOperations(opt ...OperationOption) *OperationConfig {
	config := &OperationConfig{
		MaxDepth: -1,
	}

	for _, o := range opt {
		o(config)
	}
	return config
}
