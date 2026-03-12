package cmd

import "strings"

type call struct {
	args []string
}

type mockTmux struct {
	calls    []call
	outputs  map[string]string
	attached string
}

func newMockTmux() *mockTmux {
	return &mockTmux{
		outputs: make(map[string]string),
	}
}

func (m *mockTmux) Run(args ...string) (string, error) {
	m.calls = append(m.calls, call{args: args})
	key := strings.Join(args, " ")
	if out, ok := m.outputs[key]; ok {
		return out, nil
	}
	return "", nil
}

func (m *mockTmux) Attach(sessionName string) error {
	m.attached = sessionName
	return nil
}

func (m *mockTmux) HasSession(name string) bool {
	return false
}

func (m *mockTmux) findCalls(command string) []call {
	var matched []call
	for _, c := range m.calls {
		if len(c.args) > 0 && c.args[0] == command {
			matched = append(matched, c)
		}
	}
	return matched
}

func (m *mockTmux) hasCall(args ...string) bool {
	for _, c := range m.calls {
		if len(c.args) == len(args) {
			match := true
			for i := range args {
				if args[i] != c.args[i] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}
