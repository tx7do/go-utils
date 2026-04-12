package machineid

import (
	"strings"
	"testing"
)

func TestMachineID(t *testing.T) {
	id, err := machineID()
	if err != nil {
		t.Errorf("machineID() error: %v (may be expected in restricted/CI environments)", err)
	}
	if id == "" {
		t.Error("machineID() returned empty string; this may be normal in some environments")
	} else {
		// 简单验证 ID 格式（8-4-4-4-12 的十六进制字符串）
		parts := strings.Split(id, "-")
		if len(parts) != 5 {
			t.Errorf("machineID() returned invalid format: %s", id)
		}
		t.Logf("Machine ID: %s", id)
	}
}

func TestGetMachineGuid(t *testing.T) {
	guid, err := getMachineGuid()
	if err != nil {
		t.Errorf("getMachineGuid() error: %v (may be expected on some systems)", err)
	}
	if guid == "" {
		t.Error("getMachineGuid() returned empty string; this may be normal in some environments")
	} else {
		// 简单验证 GUID 格式（8-4-4-4-12 的十六进制字符串）
		parts := strings.Split(guid, "-")
		if len(parts) != 5 {
			t.Errorf("getMachineGuid() returned invalid GUID format: %s", guid)
		}
		t.Logf("Machine GUID: %s", guid)
	}
}

func TestGetBIOSUUID(t *testing.T) {
	uuid, err := getBIOSUUID()
	if err != nil {
		t.Logf("getBIOSUUID() error: %v (wmic 可能不存在于部分系统, 可忽略)", err)
	}
	if uuid == "" {
		t.Log("getBIOSUUID() returned empty string; 这在部分系统上是正常的")
	} else {
		// 简单验证 UUID 格式（8-4-4-4-12 的十六进制字符串）
		parts := strings.Split(uuid, "-")
		if len(parts) != 5 {
			t.Errorf("getBIOSUUID() returned invalid UUID format: %s", uuid)
		}
		t.Logf("BIOS UUID: %s", uuid)
	}
}

func TestGetPrimaryMAC(t *testing.T) {
	mac, err := getPrimaryMAC()
	if err != nil {
		t.Errorf("getPrimaryMAC() error: %v (may be expected on some systems)", err)
	}
	if mac == "" {
		t.Error("getPrimaryMAC() returned empty string; this may be normal in some environments")
	} else {
		// 支持 Windows 下的 - 或 : 分隔
		var parts []string
		if strings.Contains(mac, "-") {
			parts = strings.Split(mac, "-")
		} else {
			parts = strings.Split(mac, ":")
		}
		if len(parts) != 6 {
			t.Errorf("getPrimaryMAC() returned invalid MAC address format: %s", mac)
		}
		t.Logf("Primary MAC address: %s", mac)
	}
}
