package gandalf

import (
	"fmt"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

// ScopeTree holds the parents and children for any given Scope
type ScopeTree struct {
	Root     string
	Parent   map[string]string
	Children map[string][]string
}

// ErrInvalidFormat is returned when an invalid value is encountered in the yaml file
type ErrInvalidFormat struct {
	original *interface{}
	format   string
}

func (eif ErrInvalidFormat) Error() string {
	return fmt.Sprintf("Invalid format '%s' for '%+v'", eif.format, *eif.original)
}

func newInvalidFormatErr(v interface{}) ErrInvalidFormat {
	t := reflect.TypeOf(v)
	tString := ""
	if t == nil {
		tString = "<nil>"
	} else {
		tString = t.String()
	}
	return ErrInvalidFormat{
		original: &v,
		format:   tString,
	}
}

func traverseTree(st *ScopeTree, parent string, child interface{}) error {
	var rawScope []interface{}
	var ys map[interface{}]interface{}
	var ok bool

	// child can be either a []string, a map, or nil
	// if it's nil then it's ignored
	rawScope, ok = child.([]interface{})
	if ok {
		st.Children[parent] = []string{}
		for _, scope := range rawScope {
			s, ok := scope.(string)
			if !ok {
				return newInvalidFormatErr(scope)
			}
			s = parent + "/" + s
			st.Parent[s] = parent
			st.Children[parent] = append(st.Children[parent], s)
		}
	} else {
		ys, ok = child.(map[interface{}]interface{})
		if !ok {
			if child == nil {
				return nil
			}
			return newInvalidFormatErr(child)
		}
		st.Children[parent] = []string{}
		for key, value := range ys {
			ks := key.(string)
			ks = parent + "/" + ks
			err := traverseTree(st, ks, value)
			if err != nil {
				return err
			}
			st.Parent[ks] = parent
			st.Children[parent] = append(st.Children[parent], ks)
		}
	}

	return nil
}

func toScopeTree(root string, ys map[interface{}]interface{}) (st *ScopeTree, err error) {
	st = new(ScopeTree)
	st.Parent = make(map[string]string)
	st.Children = make(map[string][]string)

	err = traverseTree(st, root, map[interface{}]interface{}(ys))

	if err != nil {
		st = nil
	}

	return
}

// MakeTree makes a new ScopeTree from YAML data
//
// root is the name of the root node.
func MakeTree(b []byte, root string) (st *ScopeTree, err error) {
	ys := make(map[interface{}]interface{})
	err = yaml.Unmarshal(b, &ys)
	if err != nil {
		return
	}

	st, err = toScopeTree(root, ys)
	if err != nil {
		return
	}
	st.Root = root

	return
}
