package linux

import (
	"git.adyxax.org/adyxax/gonf/v2/pkg"
	"os/exec"
)

func Useradd(data gonf.UserData) (gonf.Status, error) {
	getent := exec.Command("getent", "passwd", data.Name)
	if err := getent.Run(); err == nil {
		return gonf.KEPT, nil
	}
	args := make([]string, 0)
	if data.HomeDir != "" {
		args = append(args, "--home-dir="+data.HomeDir)
	}
	if data.System {
		args = append(args, "--system")
	}
	args = append(args, data.Name)
	cmd := exec.Command("useradd", args...)
	if err := cmd.Run(); err != nil {
		return gonf.BROKEN, err
	}
	return gonf.REPAIRED, nil
}
