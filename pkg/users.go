package gonf

import (
	"log/slog"
)

var users []*UserPromise

var userAddFunction func(data UserData) (Status, error)

func init() {
	users = make([]*UserPromise, 0)
}

func SetUsersConfiguration(useradd func(data UserData) (Status, error)) {
	userAddFunction = useradd
}

func User(data UserData) *UserPromise {
	if data.Name == "" {
		panic("User() promise invoked without specifying a username")
	}
	return &UserPromise{
		chain:  nil,
		data:   data,
		states: nil,
		status: PROMISED,
	}
}

type UserData struct {
	HomeDir string
	Name    string
	System  bool
}

type UserPromise struct {
	chain  []Promise
	data   UserData
	states []string
	status Status
}

func (u *UserPromise) IfRepaired(p ...Promise) Promise {
	u.chain = append(u.chain, p...)
	return u
}

func (u *UserPromise) Promise() *UserPromise {
	users = append(users, u)
	return u
}

func (u *UserPromise) Resolve() {
	var err error
	u.status, err = userAddFunction(u.data)
	switch u.status {
	case BROKEN:
		slog.Error("user", "name", u.data.Name, "status", u.status, "error", err)
	case KEPT:
		slog.Debug("user", "name", u.data.Name, "status", u.status)
	case REPAIRED:
		slog.Info("user", "name", u.data.Name, "status", u.status)
		if u.status == REPAIRED {
			for _, pp := range u.chain {
				pp.Resolve()
			}
		}
	}
}

func (u UserPromise) Status() Status {
	return u.status
}

func resolveUsers() (status Status) {
	status = KEPT
	for _, c := range users {
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
