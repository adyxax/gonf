package borg

import (
	"log/slog"
	"path/filepath"

	gonf "git.adyxax.org/adyxax/gonf/v2/pkg"
)

type BorgServer struct {
	chain   []gonf.Promise
	clients map[string][]byte // name -> publicKey
	path    string
	user    string
	status  gonf.Status
}

var borgServer *BorgServer = nil

func (b *BorgServer) IfRepaired(p ...gonf.Promise) gonf.Promise {
	b.chain = append(b.chain, p...)
	return b
}

func (b *BorgServer) Promise() *BorgServer {
	if b.status == gonf.DECLARED {
		b.status = gonf.PROMISED
		gonf.MakeCustomPromise(b).Promise()
	}
	return b
}

func (b *BorgServer) Resolve() {
	b.status = gonf.KEPT
	// Borg user
	user := gonf.User(gonf.UserData{
		HomeDir: b.path,
		Name:    "borg",
		System:  true,
	})
	user.Resolve()
	switch user.Status() {
	case gonf.BROKEN:
		b.status = gonf.BROKEN
		return
	case gonf.REPAIRED:
		b.status = gonf.REPAIRED
	}
	// borg package
	switch installBorgPackage() {
	case gonf.BROKEN:
		b.status = gonf.BROKEN
		return
	case gonf.REPAIRED:
		b.status = gonf.REPAIRED
	}
	// authorized_keys
	borgDir := gonf.ModeUserGroup(0700, "borg", "borg")
	borgRO := gonf.ModeUserGroup(0400, "borg", "borg")
	file := gonf.File(filepath.Join(b.path, ".ssh/authorized_keys")).
		DirectoriesPermissions(borgDir).
		Permissions(borgRO)
	authorizedKeys := ""
	// we sort the names so that the file contents are stable
	names := make([]string, len(b.clients)-1)
	for name := range b.clients {
		names = append(names, name)
	}
	for _, name := range names {
		key := b.clients[name]
		authorizedKeys += "command=\"borg serve --restrict-to-path " + filepath.Join(b.path, name) + "\",restrict " + string(key) + "\n"
	}
	file.Contents(authorizedKeys).Resolve()
	switch file.Status() {
	case gonf.BROKEN:
		b.status = gonf.BROKEN
		return
	case gonf.REPAIRED:
		b.status = gonf.REPAIRED
	}
}

func (b BorgServer) Status() gonf.Status {
	return b.status
}

func Server() *BorgServer {
	if borgServer == nil {
		borgServer = &BorgServer{
			chain:   nil,
			clients: make(map[string][]byte),
			path:    "/srv/borg/",
			user:    "borg",
			status:  gonf.DECLARED,
		}
	}
	return borgServer
}

func (b *BorgServer) Add(name string, publicKey []byte) *BorgServer {
	if _, ok := b.clients[name]; ok {
		slog.Debug("Duplicate name for BorgServer", "name", name)
		panic("Duplicate name for BorgServer")
	}
	b.clients[name] = publicKey
	return b
}
