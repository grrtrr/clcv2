package cmd

import (
	"fmt"

	"github.com/pkg/errors"
)

// Rune implements the pflag.Value interface (pflag has no 'rune' type)
type Rune rune

func (r Rune) Type() string   { return "rune" }
func (r Rune) String() string { return fmt.Sprintf("%q", rune(r)) }

func (r *Rune) Set(s string) error {
	var sep string
	if len(s) == 0 {
		return errors.Errorf("can not set rune from empty string")
	} else if _, err := fmt.Sscanf(`"`+s+`"`, "%q", &sep); err != nil {
		return errors.Errorf("invalid input: %q", s)
	} else {
		*r = Rune(([]rune(sep))[0])
	}
	return nil
}
