package debian

import (
	"bufio"
	"bytes"
	_ "embed"
	"log/slog"
	"os"
	"os/exec"

	"git.adyxax.org/adyxax/gonf/v2/gonf"
	"git.adyxax.org/adyxax/gonf/v2/stdlib/os/systemd"
)

var packages map[string]string

func init() {
	packages_list()
}

//go:embed sources.list
var sources_list []byte

func Promise() {
	rootRO := gonf.ModeUserGroup(0444, "root", "root")
	gonf.Default("debian-release", "stable")
	gonf.AppendVariable("debian-extra-sources", "# Extra sources")
	apt_update := gonf.Command("apt-get", "update", "-qq")
	gonf.File("/etc/apt/sources.list").Permissions(rootRO).Template(sources_list).Promise().IfRepaired(apt_update)
	gonf.SetPackagesConfiguration(packages_install, packages_list, apt_update)
	gonf.Service("opensmtpd").State("enabled", "started").Promise()
	systemd.Promise()
}

func packages_install(names []string) gonf.Status {
	allKept := true
	for _, n := range names {
		if _, ok := packages[n]; !ok {
			allKept = false
		}
	}
	if allKept {
		return gonf.KEPT
	}
	args := append([]string{"install", "-y", "--no-install-recommends"}, names...)
	cmd := gonf.CommandWithEnv([]string{"DEBIAN_FRONTEND=noninteractive", "LC_ALL=C"}, "apt-get", args...)
	cmd.Resolve()
	packages_list()
	return cmd.Status
}

func packages_list() {
	packages = make(map[string]string)
	ecmd := exec.Command("dpkg-query", "-W")
	out, err := ecmd.CombinedOutput()
	if err != nil {
		slog.Error("dpkg-query", "error", err)
		os.Exit(1)
	}
	s := bufio.NewScanner(bytes.NewReader(out))
	s.Split(bufio.ScanWords)
	for s.Scan() {
		name := s.Text()
		if !s.Scan() {
			slog.Error("dpkg-query", "error", "parsing error: no version after name")
			os.Exit(1)
		}
		packages[name] = s.Text()
	}
}
