//go:build linux
// +build linux

package machineid

import (
	"regexp"
	"strings"
)

const (
	// dbusPath is the default path for dbus machine id.
	dbusPath = "/var/lib/dbus/machine-id"
	// dbusPathEtc is the default path for dbus machine id located in /etc.
	// Some systems (like Fedora 20) only know this path.
	// Sometimes it's the other way round.
	dbusPathEtc = "/etc/machine-id"

	// For older docker versions
	cgroupPath = "/proc/self/cgroup"

	// Modern docker versions should contain this information and
	// can be used as the machine-id
	mountInfoPath = "/proc/self/mountinfo"
	// DMI product UUID (部分云主机/容器环境)
	dmiProductUUIDPath = "/sys/class/dmi/id/product_uuid"
)

// machineID returns the uuid specified at `/var/lib/dbus/machine-id` or `/etc/machine-id`.
// In case of Docker, it also checks for `/proc/self/cgroup` and `/proc/self/mountinfo`.
// If there is an error reading the files an empty string is returned.
// See https://unix.stackexchange.com/questions/144812/generate-consistent-machine-unique-id
func machineID() (string, error) {
	id, err := getFirstValidValue(
		getIDFromFile(dbusPath),
		getIDFromFile(dbusPathEtc),
		getIDFromDMI,
		getCGroup,
		getMountInfo,
	)
	if err != nil {
		return "", err
	}
	return trim(id), nil
}

// getIDFromDMI 读取 /sys/class/dmi/id/product_uuid
func getIDFromDMI() (string, error) {
	idBytes, err := readFile(dmiProductUUIDPath)
	if err != nil {
		return "", nil
	}
	id := strings.TrimSpace(string(idBytes))
	if id == "" || id == "00000000-0000-0000-0000-000000000000" {
		return "", nil
	}
	return id, nil
}

func getCGroup() (string, error) {
	cgroup, err := readFile(cgroupPath)
	if err != nil {
		return "", nil
	}

	groups := strings.Split(string(cgroup), "/")
	if len(groups) < 3 {
		return "", errors.New("cgroup is not complete")
	}

	return groups[2], nil
}

var containerIDRegex = regexp.MustCompile(`\/docker\/containers/([a-f0-9]+)/hostname`)

func getMountInfo() (string, error) {
	mountInfoBytes, err := readFile(mountInfoPath)
	if err != nil {
		return "", err
	}

	mountInfo := string(mountInfoBytes)
	if !strings.Contains(mountInfo, "docker") {
		return "", errors.New("environment is not a docker container")
	}

	foundGroups := containerIDRegex.FindStringSubmatch(mountInfo)
	if len(foundGroups) < 2 {
		return "", errors.New("no docker mountinfo found")
	}

	return foundGroups[1], nil
}
