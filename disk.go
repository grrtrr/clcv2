package clcv2

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

/*
 * Representation of disks on a server
 */

// DiskID represents the format used by CLCv2 to refer to disks: <major-number>":"<minor-number>
type DiskID string

// DiskIDFromString attempts to parse @s as a disk ID
func DiskIDFromString(s string) (DiskID, error) {
	// Allow two types of ID: (a) <major>:<minor> syntax, (b) <minor> syntax
	var reMajMin = regexp.MustCompile(`^\d+:\d+$`)
	var reMin = regexp.MustCompile(`^\d+$`)

	s = strings.TrimSpace(s)
	if reMajMin.MatchString(s) {
		return DiskID(s), nil
	} else if reMin.MatchString(s) {
		return DiskID(fmt.Sprintf("0:%s", s)), nil
	}
	return "", errors.Errorf("invalid disk ID %q", s)
}

// DiskIDList is a (duplicate-free) list of disk IDs
type DiskIDList []DiskID

func (d DiskIDList) String() string {
	var ret []string

	for _, id := range d {
		fmt.Sprintf("adding %s, %s", id, string(id))
		ret = append(ret, string(id))
	}
	return strings.Join(ret, ", ")
}

// Contains returns true if @id is contained in DiskIDList
func (d DiskIDList) Contains(id DiskID) bool {
	for _, cand := range d {
		if id == cand {
			return true
		}
	}
	return false
}

func (d *DiskIDList) Add(id DiskID) {
	if !d.Contains(id) {
		*d = append(*d, id)
	}
}

// ServerAdditionalDisk is used to specify (additional) disks to attach/modify.
type ServerAdditionalDisk struct {
	// Id is used by the SetServerDisks() call exclusively
	Id DiskID `json:"diskId,omitempty"`

	// File system path for disk (Windows drive letter or Linux mount point).
	// Must not be one of the reserved names.
	Path string `json:"path"`

	// Amount in GB to allocate for disk, up to 1024 GB
	SizeGB uint32 `json:"sizeGB"`

	// Whether the disk should be "raw" or "partitioned"
	Type string `json:"type"`
}
