package systemd

import (
	"errors"
	"os/exec"

	gonf "git.adyxax.org/adyxax/gonf/v2/pkg"
)

func Promise() {
	gonf.SetServiceFunction(systemdService)
}

func isEnabled(name string) bool {
	return systemctlShow(name, "UnitFileState") == "enabled"
}

func isRunning(name string) bool {
	return systemctlShow(name, "SubState") == "running"
}

func systemctl(name, operation string) (gonf.Status, error) {
	cmd := exec.Command("systemctl", operation, name)
	if err := cmd.Run(); err != nil {
		return gonf.BROKEN, err
	}
	return gonf.REPAIRED, nil
}

func systemctlShow(name, field string) string {
	ecmd := exec.Command("systemctl", "show", name, "-p", field, "--value")
	out, _ := ecmd.CombinedOutput()
	return string(out[:len(out)-1]) // remove trailing '\n' and convert to string
}

func systemdService(name, state string) (gonf.Status, error) {
	switch state {
	case "disabled":
		if isEnabled(name) {
			return systemctl(name, "disable")
		} else {
			return gonf.KEPT, nil
		}
	case "enabled":
		if isEnabled(name) {
			return gonf.KEPT, nil
		} else {
			return systemctl(name, "enable")
		}
	case "reloaded":
		return systemctl(name, "reloaded")
	case "restarted":
		return systemctl(name, "restart")
	case "started":
		if isRunning(name) {
			return gonf.KEPT, nil
		} else {
			return systemctl(name, "start")
		}
	case "stopped":
		if isRunning(name) {
			return systemctl(name, "stop")
		} else {
			return gonf.KEPT, nil
		}
	default:
		return gonf.BROKEN, errors.New("unsupported systemctl operation " + state)
	}
}
