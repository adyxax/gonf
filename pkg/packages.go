package gonf

import "log/slog"

var packages []*PackagePromise

var packagesInstallFunction func([]string) (Status, []string)

func init() {
	packages = make([]*PackagePromise, 0)
}

func SetPackagesConfiguration(install func([]string) (Status, []string), update *CommandPromise) {
	packagesInstallFunction = install
}

func Package(names ...string) *PackagePromise {
	return &PackagePromise{
		chain:  nil,
		err:    nil,
		names:  names,
		status: DECLARED,
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

func (p *PackagePromise) Promise() *PackagePromise {
	if p.status == DECLARED {
		p.status = PROMISED
		packages = append(packages, p)
	}
	return p
}

func (p *PackagePromise) Resolve() {
	status, affected := packagesInstallFunction(p.names)
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
