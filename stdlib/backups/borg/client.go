package borg

import (
	_ "embed"
	"fmt"
	"os"

	"log/slog"
	"path/filepath"

	gonf "git.adyxax.org/adyxax/gonf/v2/pkg"
)

//go:embed borg-script-template
var borg_script_template string

//go:embed systemd-service-template
var systemd_service_template string

//go:embed systemd-timer-template
var systemd_timer_template string

type Job struct {
	hostname   string
	path       string
	privateKey []byte
}

type BorgClient struct {
	chain  []gonf.Promise
	jobs   map[string]*Job // name -> privateKey
	path   string
	status gonf.Status
}

var borgClient *BorgClient = nil

func (b *BorgClient) IfRepaired(p ...gonf.Promise) gonf.Promise {
	b.chain = append(b.chain, p...)
	return b
}

func (b *BorgClient) Promise() *BorgClient {
	if b.status == gonf.DECLARED {
		b.status = gonf.PROMISED
		gonf.MakeCustomPromise(b).Promise()
	}
	return b
}

func (b *BorgClient) Resolve() {
	b.status = gonf.KEPT
	// borg package
	switch installBorgPackage() {
	case gonf.BROKEN:
		b.status = gonf.BROKEN
		return
	case gonf.REPAIRED:
		b.status = gonf.REPAIRED
	}
	// private key
	rootDir := gonf.ModeUserGroup(0700, "root", "root")
	rootRO := gonf.ModeUserGroup(0400, "root", "root")
	rootRX := gonf.ModeUserGroup(0500, "root", "root")
	gonf.Directory("/root/.cache/borg").DirectoriesPermissions(rootDir).Resolve()
	gonf.Directory("/root/.config/borg").DirectoriesPermissions(rootDir).Resolve()
	systemdSystemPath := "/etc/systemd/system/"
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("Unable to find current hostname", "err", err)
		hostname = "unknown"
	}
	for name, job := range b.jobs {
		gonf.File(filepath.Join(b.path, fmt.Sprintf("%s.key", name))).
			DirectoriesPermissions(rootDir).
			Permissions(rootRO).Contents(job.privateKey).
			Resolve()
		gonf.File(filepath.Join(b.path, fmt.Sprintf("%s.sh", name))).
			DirectoriesPermissions(rootDir).
			Permissions(rootRX).Contents(fmt.Sprintf(borg_script_template,
			hostname,
			name,
			job.path,
			job.hostname, name)).
			Resolve()
		service_name := fmt.Sprintf("borgbackup-job-%s.service", name)
		gonf.File(filepath.Join(systemdSystemPath, service_name)).
			DirectoriesPermissions(rootDir).
			Permissions(rootRO).Contents(fmt.Sprintf(systemd_service_template,
			name,
			job.hostname, name,
			name,
			name)).
			Resolve()
		timer_name := fmt.Sprintf("borgbackup-job-%s.timer", name)
		gonf.File(filepath.Join(systemdSystemPath, timer_name)).
			DirectoriesPermissions(rootDir).
			Permissions(rootRO).Contents(fmt.Sprintf(systemd_timer_template, name)).
			Resolve()
		gonf.Service(timer_name).State("enabled", "started").Resolve()
	}
}

func (b BorgClient) Status() gonf.Status {
	return b.status
}

func Client() *BorgClient {
	if borgClient == nil {
		borgClient = &BorgClient{
			chain:  nil,
			jobs:   make(map[string]*Job),
			path:   "/etc/borg/",
			status: gonf.DECLARED,
		}
	}
	return borgClient
}

func (b *BorgClient) Add(name string, path string, privateKey []byte, hostname string) *BorgClient {
	if _, ok := b.jobs[name]; ok {
		slog.Debug("Duplicate name for BorgClient", "name", name)
		panic("Duplicate name for BorgClient")
	}
	b.jobs[name] = &Job{
		hostname:   hostname,
		path:       path,
		privateKey: privateKey,
	}
	return b
}
