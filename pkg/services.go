package gonf

import "log/slog"

// ----- Globals ---------------------------------------------------------------
var services []*ServicePromise

// service management function
var serviceFunction func(string, string) (Status, error)

// ----- Init ------------------------------------------------------------------
func init() {
	services = make([]*ServicePromise, 0)
}

// ----- Public ----------------------------------------------------------------
func SetServiceFunction(f func(string, string) (Status, error)) {
	serviceFunction = f
}

func Service(names ...string) *ServicePromise {
	return &ServicePromise{
		chain:  nil,
		err:    nil,
		names:  names,
		states: nil,
		status: PROMISED,
	}
}

func (s *ServicePromise) State(states ...string) *ServicePromise {
	s.states = states
	return s
}

type ServicePromise struct {
	chain  []Promise
	err    error
	names  []string
	states []string
	status Status
}

func (s *ServicePromise) IfRepaired(ps ...Promise) Promise {
	s.chain = ps
	return s
}

func (s *ServicePromise) Promise() Promise {
	services = append(services, s)
	return s
}

func (s *ServicePromise) Resolve() {
	for _, name := range s.names {
		var repaired = false
		for _, state := range s.states {
			s.status, s.err = serviceFunction(name, state)
			if s.status == BROKEN {
				slog.Error("service", "name", name, "state", state, "status", s.status, "error", s.err)
				return
			} else if s.status == REPAIRED {
				repaired = true
			}
		}
		if repaired {
			s.status = REPAIRED
			slog.Info("service", "name", name, "state", s.states, "status", s.status)
		} else {
			s.status = KEPT
			slog.Debug("service", "name", name, "state", s.states, "status", s.status)
		}
	}
	if s.status == REPAIRED {
		for _, pp := range s.chain {
			pp.Resolve()
		}
	}
}

func (s ServicePromise) Status() Status {
	return s.status
}

// ----- Internal --------------------------------------------------------------
func resolveServices() (status Status) {
	status = KEPT
	for _, c := range services {
		if c.status == PROMISED {
			c.Resolve()
			switch c.status {
			case BROKEN:
				return BROKEN
			case REPAIRED:
				status = REPAIRED
			}
		}
	}
	return
}
