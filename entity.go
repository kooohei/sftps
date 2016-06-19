package sftps

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	//"github.com/davecgh/go-spew/spew"
)

type Permission struct {
	Read  bool
	Write bool
	Exe   bool
}
type Permissions struct {
	Type   string
	Sticky bool
	SUID   bool
	SGID   bool
	Owner  *Permission
	Group  *Permission
	Users  *Permission
}

type Entity struct {
	Perms   *Permissions
	Links   int
	Owner   string
	Group   string
	Size    int
	LastMod string
	Name    string
}

func StringToEntities(raw string) (ents []*Entity, err error) {
	lines := strings.Split(raw, "\r\n")

	if len(lines) == 1 {
		lines = strings.Split(raw, "\n")
	}

	for _, line := range lines {
		if line != "" {
			ent := new(Entity)
			var er error
			ent.Perms, ent.Links, ent.Owner, ent.Group, ent.Size, ent.LastMod, ent.Name, er = decomposition(line)
			if er != nil {
				err = er
				return
			}
			ents = append(ents, ent)
		}
	}
	return
}

func decomposition(line string) (perms *Permissions, link int, own string, grp string, size int, modified string, name string, err error) {
	if line == "" {
		return
	}
	reg := regexp.MustCompile(`^([dplcbs\-])([rwxstST\-]{3})([rwxstST\-]{3})([rwxstST\-]{3})\S*?\s+?(\d+?)\s+?(\S+?)\s+?(\S+?)\s+?(\d+?)\s+?(\S+?\s+?\d+?\s+?(?:\d{2}\:\d{2}|\d{4}))\s+?(.+?)$`)
	cols := reg.FindStringSubmatch(line)
	if cols == nil {
		return
	}
	if len(cols) == 0 {
		err = errors.New("Failed to parse the unix file format.")
		return
	}
	cols = cols[1:]
	if len(cols) != 10 {
		err = errors.New("Could not parse to json, passed string is unknown format.")
	}

	perms = getPermissions(cols[0], cols[1], cols[2], cols[3])
	link, err = strconv.Atoi(cols[4])
	if err != nil {
		return
	}

	own = cols[5]
	grp = cols[6]
	size, err = strconv.Atoi(cols[7])
	if err != nil {
		return
	}
	modified = cols[8]
	name = cols[9]

	return
}

func getPermissions(tp string, own string, grp string, usr string) (res *Permissions) {
	res = new(Permissions)
	res.Type = getFileTypeFromChar(tp)
	res.Sticky = (usr[2] == 't' || usr[2] == 'T')
	res.SUID = (usr[2] == 's' || usr[2] == 'S')
	res.SGID = (grp[2] == 's' || grp[2] == 'G')
	res.Users = getPermission(usr)
	res.Group = getPermission(grp)
	res.Owner = getPermission(own)
	return
}

func getPermission(perm string) (res *Permission) {
	res = new(Permission)
	res.Read = perm[0] == 'r'
	res.Write = perm[1] == 'w'
	res.Exe = (perm[2] == 'x' || perm[2] == 's')

	return
}

func getFileTypeFromChar(ch string) (tp string) {
	if ch == "d" {
		tp = "Directory"
	} else if ch == "-" {
		tp = "Regular"
	} else if ch == "l" {
		tp = "Symlink"
	} else if ch == "p" {
		tp = "Pipe"
	} else if ch == "s" {
		tp = "Socket"
	} else if ch == "c" {
		tp = "CharacterDevice"
	} else if ch == "b" {
		tp = "BlockDevice"
	}
	return
}