package form

import "github.com/asaskevich/govalidator"

type Complexity struct {
	MinLength int
	MaxLength int
	Patterns  []string
}

func (c *Complexity) Validate(s string) bool {
	if !govalidator.IsByteLength(s, c.MinLength, c.MaxLength) {
		return false
	}
	for _, p := range c.Patterns {
		if !govalidator.Matches(s, p) {
			return false
		}
	}
	return true
}
