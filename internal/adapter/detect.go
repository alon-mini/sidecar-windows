package adapter

// adapterFactories holds registered adapter constructors.
var adapterFactories []func() Adapter

// RegisterFactory registers an adapter constructor.
func RegisterFactory(factory func() Adapter) {
	adapterFactories = append(adapterFactories, factory)
}

// DetectAdapters scans for available adapters for the given project.
func DetectAdapters(projectRoot string) (map[string]Adapter, error) {
	adapters := make(map[string]Adapter)
	for _, factory := range adapterFactories {
		instance := factory()
		detected, err := instance.Detect(projectRoot)
		if err != nil || !detected {
			continue
		}
		adapters[instance.ID()] = instance
	}
	return adapters, nil
}

// AllAdapters creates all registered adapter instances without filtering by Detect.
// Use this at startup so that all adapters are available across project switches;
// consumers (e.g. conversations plugin) call Detect() per-adapter to filter by project.
func AllAdapters() map[string]Adapter {
	adapters := make(map[string]Adapter, len(adapterFactories))
	for _, factory := range adapterFactories {
		instance := factory()
		adapters[instance.ID()] = instance
	}
	return adapters
}
