package auditutil

import (
	"regexp"
	"strings"
)

// 平台类型常量（DetectPlatformFromUA 的取值域）。
const (
	PlatformAndroidApp = "AndroidApp"
	PlatformiOSApp     = "iOSApp"

	PlatformDesktopWindows = "DesktopWindows"
	PlatformDesktopMac     = "DesktopMac"
	PlatformDesktopLinux   = "DesktopLinux"

	PlatformWeb = "Web"

	PlatformOther = "Other"
)

var reAndroidPackage = regexp.MustCompile(`\bcom\.[a-z0-9_.]+`)

// DetectPlatformFromUA 根据 UA 字符串启发式判断平台类型。
// 说明：仅作为补充判断，优先使用客户端上报字段（platform/os_name）。
func DetectPlatformFromUA(ua string) string {
	if ua == "" {
		return PlatformOther
	}
	s := strings.ToLower(strings.TrimSpace(ua))

	// 原生 Android app 指示器
	if strings.Contains(s, "okhttp") || strings.Contains(s, "dalvik") || strings.Contains(s, "; wv") ||
		strings.Contains(s, " ;wv") || strings.Contains(s, "build/") {
		return PlatformAndroidApp
	}
	// 包名模式（com.xxx）且含 android
	if strings.Contains(s, "android") {
		if reAndroidPackage.MatchString(s) || strings.Contains(s, "wv") {
			return PlatformAndroidApp
		}
	}

	// 原生 iOS 指示器
	if strings.Contains(s, "iphone") || strings.Contains(s, "ipad") || strings.Contains(s, "ipod") ||
		strings.Contains(s, "cfnetwork") || strings.Contains(s, "darwin") || strings.Contains(s, "cpu iphone os") {
		return PlatformiOSApp
	}

	// 桌面原生/混合应用（例如 Electron、nwjs、desktop）
	if strings.Contains(s, "electron") || strings.Contains(s, "nwjs") || strings.Contains(s, "node.js") || strings.Contains(s, "nodejs") ||
		strings.Contains(s, "desktop") || strings.Contains(s, "appname") {
		// 根据 UA 中的 OS 关键词区分具体桌面系统
		if strings.Contains(s, "windows nt") || strings.Contains(s, "win64") || strings.Contains(s, "win32") || strings.Contains(s, "windows") {
			return PlatformDesktopWindows
		}
		if strings.Contains(s, "macintosh") || strings.Contains(s, "mac os x") || strings.Contains(s, "darwin") {
			return PlatformDesktopMac
		}
		if strings.Contains(s, "x11") || strings.Contains(s, "linux") || strings.Contains(s, "ubuntu") || strings.Contains(s, "debian") || strings.Contains(s, "fedora") {
			return PlatformDesktopLinux
		}
		// 若未直接包含 OS 关键字，尝试从常见标识推断
		if strings.Contains(s, "win") || strings.Contains(s, "windows") {
			return PlatformDesktopWindows
		}
		if strings.Contains(s, "mac") || strings.Contains(s, "os x") {
			return PlatformDesktopMac
		}
		if strings.Contains(s, "linux") || strings.Contains(s, "x11") {
			return PlatformDesktopLinux
		}
		// 仍无法判定，视作桌面类但未知系统
		return PlatformOther
	}

	// 常见 Web 浏览器识别（Mozilla+桌面/移动系统且无明显原生 app 标志）
	if strings.Contains(s, "mozilla") && (strings.Contains(s, "windows nt") || strings.Contains(s, "macintosh") ||
		strings.Contains(s, "x11") || strings.Contains(s, "linux") || strings.Contains(s, "android") || strings.Contains(s, "iphone")) {
		// 排除明显的原生 app 标志
		if !strings.Contains(s, "okhttp") && !strings.Contains(s, "dalvik") && !strings.Contains(s, "cfnetwork") && !strings.Contains(s, "electron") {
			return PlatformWeb
		}
	}

	return PlatformOther
}
