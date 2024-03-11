package gonf

import (
	"errors"
	"io/fs"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

type Permissions struct {
	group Value
	mode  Value
	user  Value
}

func ModeUserGroup(mode, user, group interface{}) *Permissions {
	return &Permissions{
		group: interfaceToTemplateValue(group),
		mode:  interfaceToTemplateValue(mode),
		user:  interfaceToTemplateValue(user),
	}
}

func (p *Permissions) resolve(filename string) (Status, error) {
	g, ok := p.group.(*IntValue)
	if !ok {
		if group, err := user.LookupGroup(p.group.String()); err != nil {
			return BROKEN, err
		} else {
			if groupId, err := strconv.Atoi(group.Gid); err != nil {
				return BROKEN, err
			} else {
				g = &IntValue{groupId}
				p.group = g
			}
		}
	}
	m, err := p.mode.Int()
	if err != nil {
		return BROKEN, err
	}
	u, ok := p.user.(*IntValue)
	if !ok {
		if user, err := user.Lookup(p.user.String()); err != nil {
			return BROKEN, err
		} else {
			if userId, err := strconv.Atoi(user.Uid); err != nil {
				return BROKEN, err
			} else {
				u = &IntValue{userId}
				p.group = u
			}
		}
	}
	var status Status = KEPT
	if fileInfo, err := os.Lstat(filename); err != nil {
		return BROKEN, err
	} else {
		gv, _ := g.Int()
		mv := fs.FileMode(m)
		uv, _ := u.Int()
		if fileInfo.Mode() != mv {
			if err := os.Chmod(filename, mv); err != nil {
				return BROKEN, err
			}
			status = REPAIRED
		}
		if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			if stat.Gid != uint32(gv) || stat.Uid != uint32(uv) {
				if err := os.Chown(filename, uv, gv); err != nil {
					return BROKEN, err
				}
				status = REPAIRED
			}
		} else {
			return BROKEN, errors.New("Unsupported operating system")
		}
	}
	return status, nil
}
