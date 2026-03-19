package cmd

import "strings"

type call struct {
	args []string
}

type mockTmux struct {
	calls      []call
	outputs    map[string]string
	outputSeqs map[string][]string // sequential outputs for repeated calls
	seqIndex   map[string]int
	attached   string
}

func newMockTmux() *mockTmux {
	return &mockTmux{
		outputs:    make(map[string]string),
		outputSeqs: make(map[string][]string),
		seqIndex:   make(map[string]int),
	}
}

func (m *mockTmux) Run(args ...string) (string, error) {
	m.calls = append(m.calls, call{args: args})
	key := strings.Join(args, " ")
	if seq, ok := m.outputSeqs[key]; ok {
		i := m.seqIndex[key]
		if i < len(seq) {
			m.seqIndex[key] = i + 1
			return seq[i], nil
		}
	}
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
