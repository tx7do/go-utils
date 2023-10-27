package ioutil

import (
	"os"
	"path/filepath"
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

// ReadFile 读取文件
func ReadFile(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return content
}
