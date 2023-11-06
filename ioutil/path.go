package ioutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

// GetWorkingDirPath 获取工作路径
func GetWorkingDirPath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// GetExePath 获取可执行程序路径
func GetExePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	return exePath
}

// GetAbsPath 获取绝对路径
func GetAbsPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	return dir
}

// GetFileList 获取文件夹下面的所有文件的列表
func GetFileList(root string) []string {
	var files []string

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	}); err != nil {
		return nil
	}

	return files
}

// GetFolderNameList 获取当前文件夹下面的所有文件夹名的列表
func GetFolderNameList(root string) []string {
	var names []string
	fs, _ := os.ReadDir(root)
	for _, file := range fs {
		if file.IsDir() {
			names = append(names, file.Name())
		}
	}
	return names
}

// MatchPath Returns whether a given path matches a glob pattern.
//
// via github.com/gobwas/glob:
//
// Compile creates Glob for given pattern and strings (if any present after pattern) as separators.
// The pattern syntax is:
//
//	pattern:
//	    { term }
//
//	term:
//	    `*`         matches any sequence of non-separator characters
//	    `**`        matches any sequence of characters
//	    `?`         matches any single non-separator character
//	    `[` [ `!` ] { character-range } `]`
//	                character class (must be non-empty)
//	    `{` pattern-list `}`
//	                pattern alternatives
//	    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
//	    `\` c       matches character c
//
//	character-range:
//	    c           matches character c (c != `\\`, `-`, `]`)
//	    `\` c       matches character c
//	    lo `-` hi   matches character c for lo <= c <= hi
//
//	pattern-list:
//	    pattern { `,` pattern }
//	                comma-separated (without spaces) patterns
func MatchPath(pattern string, path string) bool {
	if g, err := glob.Compile(pattern); err == nil {
		return g.Match(path)
	}

	return false
}

// ExpandUser replaces the tilde (~) in a path into the current user's home directory.
func ExpandUser(path string) (string, error) {
	if u, err := user.Current(); err == nil {
		fullTilde := fmt.Sprintf("~%s", u.Name)

		if strings.HasPrefix(path, `~/`) || path == `~` {
			return strings.Replace(path, `~`, u.HomeDir, 1), nil
		}

		if strings.HasPrefix(path, fullTilde+`/`) || path == fullTilde {
			return strings.Replace(path, fullTilde, u.HomeDir, 1), nil
		}

		return path, nil
	} else {
		return path, err
	}
}

// IsNonemptyExecutableFile Returns true if the given path is a regular file, is executable by any user, and has a non-zero size.
func IsNonemptyExecutableFile(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.Size() > 0 && (stat.Mode().Perm()&0111) != 0 {
		return true
	}

	return false
}

// IsNonemptyFile Returns true if the given path is a regular file with a non-zero size.
func IsNonemptyFile(path string) bool {
	if FileExists(path) {
		if stat, err := os.Stat(path); err == nil && stat.Size() > 0 {
			return true
		}
	}

	return false
}

// IsNonemptyDir Returns true if the given path is a directory with items in it.
func IsNonemptyDir(path string) bool {
	if DirExists(path) {
		if entries, err := ioutil.ReadDir(path); err == nil && len(entries) > 0 {
			return true
		}
	}

	return false
}

// Exists Returns true if the given path exists.
func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

// LinkExists Returns true if the given path exists and is a symbolic link.
func LinkExists(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return IsSymlink(stat.Mode())
	}

	return false
}

// FileExists Returns true if the given path exists and is a regular file.
func FileExists(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return stat.Mode().IsRegular()
	}

	return false
}

// DirExists Returns true if the given path exists and is a directory.
func DirExists(path string) bool {
	if stat, err := os.Stat(path); err == nil {
		return stat.IsDir()
	}

	return false
}

// PathExist 路径是否存在
func PathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func IsSymlink(mode os.FileMode) bool {
	return mode&os.ModeSymlink != 0
}

func IsDevice(mode os.FileMode) bool {
	return mode&os.ModeDevice != 0
}

func IsCharDevice(mode os.FileMode) bool {
	return mode&os.ModeCharDevice != 0
}

func IsNamedPipe(mode os.FileMode) bool {
	return mode&os.ModeNamedPipe != 0
}

func IsSocket(mode os.FileMode) bool {
	return mode&os.ModeSocket != 0
}

func IsSticky(mode os.FileMode) bool {
	return mode&os.ModeSticky != 0
}

func IsSetuid(mode os.FileMode) bool {
	return mode&os.ModeSetuid != 0
}

func IsSetgid(mode os.FileMode) bool {
	return mode&os.ModeSetgid != 0
}

func IsTemporary(mode os.FileMode) bool {
	return mode&os.ModeTemporary != 0
}

func IsExclusive(mode os.FileMode) bool {
	return mode&os.ModeExclusive != 0
}

func IsAppend(mode os.FileMode) bool {
	return mode&os.ModeAppend != 0
}

// IsReadable Returns true if the given file can be opened for reading by the current user.
func IsReadable(filename string) bool {
	if f, err := os.OpenFile(filename, os.O_RDONLY, 0); err == nil {
		defer f.Close()
		return true
	} else {
		return false
	}
}

// IsWritable Returns true if the given file can be opened for writing by the current user.
func IsWritable(filename string) bool {
	if f, err := os.OpenFile(filename, os.O_WRONLY, 0); err == nil {
		defer f.Close()
		return true
	} else {
		return false
	}
}

// IsAppendable Returns true if the given file can be opened for appending by the current user.
func IsAppendable(filename string) bool {
	if f, err := os.OpenFile(filename, os.O_APPEND, 0); err == nil {
		defer f.Close()
		return true
	} else {
		return false
	}
}
