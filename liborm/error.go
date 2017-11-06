package liborm

import (
	"fmt"
)

var (
	DidnotHavePrimeKeyError = fmt.Errorf("model did not have a prime key")
	ConditionValError       = fmt.Errorf("model did not have a right filed name or key")
)
