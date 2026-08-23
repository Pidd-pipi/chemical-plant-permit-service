package validation

import "fmt"

var allowed = map[string]bool{"draft": true, "review": true, "approved": true, "suspended": true, "closed": true}

func Status(v string) error {
	if v == "" {
		return fmt.Errorf("status required")
	}
	if !allowed[v] {
		return fmt.Errorf("unsupported status %q", v)
	}
	return nil
}
