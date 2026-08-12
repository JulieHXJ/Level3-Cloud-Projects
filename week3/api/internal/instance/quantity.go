package instance

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

// cpu and storage validation, > 0
func normalizePositiveQuantity(value, field string) (string, error) {
	value = strings.TrimSpace(value)

	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return "", fmt.Errorf(
			"%w: invalid %s value %q",
			ErrInvalidInstance,
			field,
			value,
		)
	}

	if quantity.Sign() <= 0 {
		return "", fmt.Errorf(
			"%w: %s must be greater than zero",
			ErrInvalidInstance,
			field,
		)
	}

	return value, nil
}

// compare storage and forbib to scale down, only scale up
func validateStorageExpansion(current, requested string) error {
	currentQuantity, err := resource.ParseQuantity(current)
	if err != nil {
		return fmt.Errorf(
			"%w: invalid current storage value %q",
			ErrInvalidInstance,
			current,
		)
	}

	requestedQuantity, err := resource.ParseQuantity(requested)
	if err != nil {
		return fmt.Errorf(
			"%w: invalid storage value %q",
			ErrInvalidInstance,
			requested,
		)
	}

	if requestedQuantity.Cmp(currentQuantity) < 0 {
		return fmt.Errorf(
			"%w: storage cannot be decreased from %s to %s",
			ErrInvalidInstance,
			current,
			requested,
		)
	}

	return nil
}
