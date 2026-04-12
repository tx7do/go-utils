//go:build darwin
// +build darwin

package machineid

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// machineID returns the uuid returned by `ioreg -rd1 -c IOPlatformExpertDevice`.
// If there is an error running the commad an empty string is returned.
func machineID() (string, error) {
	var (
		buf bytes.Buffer
		err error
		id  string
	)
	// 优先尝试完整路径
	err = run(&buf, os.Stderr, "/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	if err != nil {
		// 回退到 PATH 中的 ioreg
		buf.Reset()
		err2 := run(&buf, os.Stderr, "ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		if err2 != nil {
			return "", fmt.Errorf("failed to run /usr/sbin/ioreg: %v; fallback to ioreg also failed: %v", err, err2)
		}
	}
	id, err = extractID(buf.String())
	if err != nil {
		return "", fmt.Errorf("failed to extract IOPlatformUUID: %v", err)
	}
	return trim(id), nil
}

func extractID(lines string) (string, error) {
	for _, line := range strings.Split(lines, "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.SplitAfter(line, `" = "`)
			if len(parts) == 2 {
				return strings.TrimRight(parts[1], `"`), nil
			}
		}
	}
	return "", fmt.Errorf("Failed to extract 'IOPlatformUUID' value from `ioreg` output.\n%s", lines)
}
