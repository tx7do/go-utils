//go:build windows
// +build windows

package machineid

import (
	"golang.org/x/sys/windows/registry"
)

// machineID returns the key MachineGuid in registry `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`.
// If there is an error running the commad an empty string is returned.
import (
	"os/exec"
	"strings"
)

func machineID() (string, error) {
	// 1. 注册表 MachineGuid
	guid, err := getMachineGuid()
	if err == nil && guid != "" {
		return guid, nil
	}

	// 2. BIOS UUID (wmic csproduct get uuid)
	uuid, err := getBIOSUUID()
	if err == nil && uuid != "" {
		return uuid, nil
	}

	// 3. 主网卡 MAC 地址
	mac, err := getPrimaryMAC()
	if err == nil && mac != "" {
		return mac, nil
	}

	return "", err
}

// getMachineGuid 读取注册表 MachineGuid
func getMachineGuid() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		return "", err
	}
	defer k.Close()
	s, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "", err
	}
	return s, nil
}

// getBIOSUUID 通过 wmic 或 PowerShell 获取 BIOS UUID
func getBIOSUUID() (string, error) {
	// 1. 先尝试 wmic
	out, err := exec.Command("wmic", "csproduct", "get", "uuid").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.EqualFold(line, "UUID") {
				return line, nil
			}
		}
	}
	// 2. 回退到 PowerShell
	psCmd := "Get-WmiObject -Class Win32_ComputerSystemProduct | Select-Object -ExpandProperty UUID"
	out, err = exec.Command("powershell", "-Command", psCmd).Output()
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(string(out))
	if uuid == "" || strings.EqualFold(uuid, "UUID") {
		return "", nil
	}
	return uuid, nil
}

// getPrimaryMAC 获取主网卡 MAC 地址
func getPrimaryMAC() (string, error) {
	out, err := exec.Command("getmac").Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && strings.Contains(fields[0], "-") {
			return fields[0], nil
		}
	}
	return "", nil
}
