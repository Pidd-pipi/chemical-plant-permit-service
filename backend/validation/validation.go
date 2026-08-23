package validation

import "fmt"

var allowed = map[string]bool{"draft": true, "review": true, "approved": true, "suspended": true, "closed": true}

func Status(v string) error {
	if !allowed[v] && v != "" {
		return fmt.Errorf("unsupported status %q", v)
	}
	return nil
}
