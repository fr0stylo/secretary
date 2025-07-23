package runtime

import "time"

type SecretRetrieverOpts = func(*Options)

/*
Options - Structure to store any options for the secretary runtime
*/
type Options struct {
	Frequency time.Duration
	Timeout   time.Duration
}

/*
WithFrequency -
*/
func WithFrequency(frequency time.Duration) SecretRetrieverOpts {
	return func(config *Options) {
		config.Frequency = frequency
	}
}

/*
WithTimeout -
*/
func WithTimeout(timeout time.Duration) SecretRetrieverOpts {
	return func(config *Options) {
		config.Timeout = timeout
	}
}

/*
DefaultOptions -
*/
func DefaultOptions() *Options {
	return &Options{
		Frequency: 15 * time.Second,
		Timeout:   10 * time.Second,
	}
}
