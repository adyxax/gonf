package gonf

import (
	"log/slog"
)

// ----- Globals ---------------------------------------------------------------
var users []*UserPromise

// users management functions
var user_add_function func(data UserData) (Status, error)

// ----- Init ------------------------------------------------------------------
func init() {
	users = make([]*UserPromise, 0)
}

// ----- Public ----------------------------------------------------------------
func SetUsersConfiguration(useradd func(data UserData) (Status, error)) {
	user_add_function = useradd
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

func (u *UserPromise) Promise() Promise {
	users = append(users, u)
	return u
}

func (u *UserPromise) Resolve() {
	var err error
	u.status, err = user_add_function(u.data)
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

// ----- Internal --------------------------------------------------------------
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
