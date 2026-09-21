// Package core defines domain models, states, and secondary port interfaces for taskrunner-go.
package core

import "errors"

var ErrJobNotFound = errors.New("job not found")
