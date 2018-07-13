package gandalf

import (
	"regexp"
	"strings"
)

var rx = regexp.MustCompile(`( )+`)

// Need checks if the user has a scope equivalent to the required scope.
func (st *ScopeTree) Need(required, scopes string) bool {
	if scopes == "*" {
		return true
	}
	scopes = strings.TrimSpace(string(rx.ReplaceAll([]byte(scopes), []byte(" "))))

	// a set data structure. struct{}'s take 0 bytes
	set := make(map[string]struct{})
	for _, scope := range strings.Split(scopes, " ") {
		set[scope] = struct{}{}
	}

	// climbs up in the tree
	components := strings.Split(st.Root+"/"+required, "/")
	for i := len(components); i > 0; i-- {
		newRequired := strings.Join(components[0:i], "/")
		if _, ok := set[newRequired]; ok {
			return true
		}
	}

	return false
}
