package gonf

import "log/slog"

// ----- Globals ---------------------------------------------------------------
var packages []*PackagePromise

// packages management functions
var packages_install_function func([]string) (Status, []string)
var packages_update_function *CommandPromise

// ----- Init ------------------------------------------------------------------
func init() {
	packages = make([]*PackagePromise, 0)
}

// ----- Public ----------------------------------------------------------------
func SetPackagesConfiguration(install func([]string) (Status, []string), update *CommandPromise) {
	packages_install_function = install
	packages_update_function = update
}

func Package(names ...string) *PackagePromise {
	return &PackagePromise{
		chain:  nil,
		err:    nil,
		names:  names,
		status: PROMISED,
	}
}

type PackagePromise struct {
	chain  []Promise
	err    error
	names  []string
	status Status
}

func (p *PackagePromise) IfRepaired(ps ...Promise) Promise {
	p.chain = append(p.chain, ps...)
	return p
}

func (p *PackagePromise) Promise() Promise {
	packages = append(packages, p)
	return p
}

func (p *PackagePromise) Resolve() {
	status, affected := packages_install_function(p.names)
	switch status {
	case BROKEN:
		slog.Error("package", "names", p.names, "status", status, "broke", affected)
	case KEPT:
		slog.Debug("package", "names", p.names, "status", status)
	case REPAIRED:
		slog.Info("package", "names", p.names, "status", status, "repaired", affected)
		for _, pp := range p.chain {
			pp.Resolve()
		}
	}
}

func (p PackagePromise) Status() Status {
	return p.status
}

// ----- Internal --------------------------------------------------------------
func resolvePackages() (status Status) {
	status = KEPT
	for _, c := range packages {
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
