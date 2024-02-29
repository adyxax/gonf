package debian

import (
	"bufio"
	"bytes"
	_ "embed"
	"log/slog"
	"os"
	"os/exec"
	"strings"

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

func packages_install(names []string) (gonf.Status, []string) {
	gonf.FilterSlice(&names, func(n string) bool {
		_, ok := packages[n]
		return !ok
	})
	if len(names) == 0 {
		return gonf.KEPT, nil
	}
	args := append([]string{"install", "-y", "--no-install-recommends"}, names...)
	cmd := gonf.CommandWithEnv([]string{"DEBIAN_FRONTEND=noninteractive", "LC_ALL=C"}, "apt-get", args...)
	cmd.Resolve()
	packages_list()
	return cmd.Status, names
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
		if strings.Contains(name, ":") { // some packages are named with the arch like something:amd64
			name = strings.Split(name, ":")[0] // in this case we want only the name
		}
		packages[name] = s.Text()
	}
}
