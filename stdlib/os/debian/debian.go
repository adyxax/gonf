package debian

import (
	_ "embed"

	gonf "git.adyxax.org/adyxax/gonf/v2/pkg"
	"git.adyxax.org/adyxax/gonf/v2/stdlib/os/linux"
	"git.adyxax.org/adyxax/gonf/v2/stdlib/os/systemd"
)

//go:embed apt-norecommends
var aptNoRecommends []byte

//go:embed sources.list
var sourcesList []byte

func Promise() {
	// ----- gonf --------------------------------------------------------------
	aptUpdate := gonf.Command("apt-get", "update", "-qq")
	gonf.SetPackagesConfiguration(packagesInstall, aptUpdate)
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
		Template(sourcesList).
		Promise().
		IfRepaired(aptUpdate)
	gonf.File("/etc/apt/apt.conf.d/99_norecommends").
		DirectoriesPermissions(rootDir).
		Permissions(rootRO).
		Contents(aptNoRecommends).
		Promise()
}
