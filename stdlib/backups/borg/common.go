package borg

import gonf "git.adyxax.org/adyxax/gonf/pkg"

func installBorgPackage() gonf.Status {
	packag := gonf.Package("borgbackup")
	packag.Resolve()
	return packag.Status()
}
