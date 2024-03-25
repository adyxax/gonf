package debian

import (
	_ "embed"
	"git.adyxax.org/adyxax/gonf/v2/pkg"
	"git.adyxax.org/adyxax/gonf/v2/stdlib/os/linux"
	"git.adyxax.org/adyxax/gonf/v2/stdlib/os/systemd"
)

//go:embed apt-norecommends
var apt_norecommends []byte

//go:embed sources.list
var sources_list []byte

func Promise() {
	// ----- gonf --------------------------------------------------------------
	apt_update := gonf.Command("apt-get", "update", "-qq")
	gonf.SetPackagesConfiguration(packages_install, apt_update)
	gonf.SetUsersConfiguration(linux.Useradd)
	// ----- systemd -----------------------------------------------------------
	systemd.Promise()
	// ----- apt ---------------------------------------------------------------
	rootDir := gonf.ModeUserGroup(0755, "root", "root")
	rootRO := gonf.ModeUserGroup(0444, "root", "root")
	gonf.Default("debian-release", "stable")
	gonf.AppendVariable("debian-extra-sources", "# Extra sources")
	gonf.File("/etc/apt/sources.list").
		Permissions(rootRO).
		Template(sources_list).
		Promise().
		IfRepaired(apt_update)
	gonf.File("/etc/apt/apt.conf.d/99_norecommends").
		DirectoriesPermissions(rootDir).
		Permissions(rootRO).
		Contents(apt_norecommends).
		Promise()
}
