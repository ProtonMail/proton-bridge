package pmapi

const defaultPageSize = 100

func doPaged(elements []string, pageSize int, fn func([]string) error) error {
	for len(elements) > pageSize {
		if err := fn(elements[:pageSize]); err != nil {
			return err
		}

		elements = elements[pageSize:]
	}

	return fn(elements)
}
