package google

type Option func(*options)

type options struct {
	version string
	apiKey  string
}

func WithVersion(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

func WithApiKey(key string) Option {
	return func(o *options) {
		o.apiKey = key
	}
}
