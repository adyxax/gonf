package gonf

// ----- Globals ---------------------------------------------------------------
var packages []*PackagePromise

// packages management functions
var packages_install_function func([]string) Status
var packages_list_function func()
var packages_update_function *CommandPromise

// ----- Init ------------------------------------------------------------------
func init() {
	packages = make([]*PackagePromise, 0)
}

// ----- Public ----------------------------------------------------------------
func SetPackagesConfiguration(install func([]string) Status, list func(), update *CommandPromise) {
	packages_install_function = install
	packages_list_function = list
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
	p.chain = ps
	return p
}

func (p *PackagePromise) Promise() Promise {
	packages = append(packages, p)
	return p
}

func (p *PackagePromise) Resolve() {
	status := packages_install_function(p.names)
	if status == REPAIRED {
		for _, pp := range p.chain {
			pp.Resolve()
		}
	}
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
