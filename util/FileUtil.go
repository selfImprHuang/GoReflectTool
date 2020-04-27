/*
 *  @Author : huangzj
 *  @Time : 2020/4/26 14:56
 *  @Description： 文件工具,主要处理权限、文件创建等操作
 *
 *
 *  打开文件的权限相关：
 *  os.O_RDONLY // 只读
 *  os.O_WRONLY // 只写
 *  os.O_RDWR // 读写
 *  os.O_APPEND // 往文件中添建（Append）
 *  os.O_CREATE // 如果文件不存在则先创建
 *  os.O_TRUNC // 文件打开时裁剪文件
 *  os.O_EXCL // 和O_CREATE一起使用，文件不能存在
 *  os.O_SYNC // 以同步I/O的方式打开
 *
 */

package util

import (
	err2 "Go-Tool/err"
	"fmt"
	"io/ioutil"
	"os"
)

/*
 * @param path 文件夹目录
 * @return err 如果报错返回错误信息，否则返回nil
 * @return bool 返回是否文件夹结果
 * @return os.FileInfo 返回FileInfo对象，如果没有的话返回nil
 * @description
 */
func IsDirExist(path string) (*err2.FileError, bool, os.FileInfo) {
	dir, err := os.Stat(path)
	//如果路仅错误直接返回错误
	if err != nil {
		return err2.ENewFileError(err), false, nil
	}
	//如果是文件夹直接返回
	if dir.IsDir() {
		return nil, true, dir
	}
	return err2.NewFileError("非文件夹"), false, nil
}

func IsFileExist(path string) (*err2.FileError, bool, os.FileInfo) {
	file, err := os.Stat(path)
	//如果路仅错误直接返回错误
	if err != nil {
		return err2.ENewFileError(err), false, nil
	}

	if file.IsDir() {
		return err2.NewFileError("非文件"), false, nil
	}

	return nil, true, file
}

/*
 * @param filepath 文件目录
 * @return err 如果报错返回错误信息，否则返回nil
 * @return bool 是否有读权限
 * @return *File 返回文件对象，如果没有返回nil
 * @description 这边需要注意的是，在打开文件的时候，如果报错，分两种情况，一个是没权限，一个是文件有问题
 */
func HasReadPermission(filePath string) (*err2.FileError, bool, *os.File) {
	eErr, exist, _ := IsFileExist(filePath)
	if eErr != nil || !exist {
		return err2.ENewFileError(eErr), false, nil
	}
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		//如果是没权限的错误返回
		if os.IsPermission(err) {
			return err2.ENewFileError(err), false, nil
		}
		//如果有权限但是发生错误的话，返回错误和有权限的标识
		return err2.ENewFileError(err), true, nil
	}

	return nil, true, file
}

/*
 * @param filepath 文件目录
 * @return err 如果报错返回错误信息，否则返回nil
 * @return bool 是否有写权限
 * @return *File 返回文件对象，如果没有返回nil
 * @description 这边需要注意的是，在打开文件的时候，如果报错，分两种情况，一个是没权限，一个是文件有问题
 */
func HasWritePermission(filePath string) (*err2.FileError, bool, *os.File) {
	eErr, exist, _ := IsFileExist(filePath)
	if eErr != nil || !exist {
		return err2.ENewFileError(eErr), false, nil
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0666)
	if err != nil {
		//如果是没权限的错误返回
		if os.IsPermission(err) {
			return err2.ENewFileError(err), false, nil
		}
		//如果有权限但是发生错误的话，返回错误和有权限的标识
		return err2.ENewFileError(err), true, nil
	}

	return nil, true, file
}

/*
 * @param filepath 文件目录
 * @return err 如果报错返回错误信息，否则返回nil
 * @return bool 是否有写权限
 * @return *File 返回文件对象，如果没有返回nil
 * @description 判断对文件是否有读写权限.
 */
func HasReadWritePermission(filePath string) (*err2.FileError, bool, *os.File) {
	rErr, hasR, rFile := HasReadPermission(filePath)
	wErr, hasW, _ := HasWritePermission(filePath)
	if hasR && hasW {
		return nil, true, rFile
	}

	if rErr != nil {
		return err2.ENewFileError(rErr), false, nil
	}

	return err2.ENewFileError(wErr), false, nil
}

/*
 * @param filePath 文件路径
 * @return FileError 文件错误
 * @return File文件指针对象
 * @description 创建文件，如果文件存在不进行覆盖
 */
func CreateFile(filePath string) (*err2.FileError, *os.File) {
	file, err := os.Create(filePath)
	if err != nil {
		return err2.ENewFileError(err), nil
	}
	return nil, file
}

/*
 * @param filePath 文件路径
 * @return FileError 文件错误
 * @return File文件指针对象
 * @description 创建文件，如果文件存在进行覆盖
 */
func CreateOrCoverFile(filePath string) (*err2.FileError, *os.File) {
	_, has, _ := IsFileExist(filePath)

	if has {
		rErr := os.Remove(filePath)
		if rErr != nil {
			return err2.ENewFileError(rErr), nil
		}
		return CreateFile(filePath)
	}
	return CreateFile(filePath)
}

/*
 * @param filePath 文件路径
 * @return File文件指针对象
 * @description 获取文件夹下面的所有文件对象数组
 */
func GetAllFileFromDir(filePath string) []*os.File {
	return getFiles(filePath)
}

/*
 * @param
 * @return
 * @description 删除文件夹下面所有的空文件夹
 */
func DeleteEmptyDir(filePath string) *err2.FileError {
	err, is, _ := IsDirExist(filePath)
	if err != nil {
		return err2.ENewFileError(err)
	}
	if !is {
		return err2.NewFileError("非文件夹")
	}

	//拿到所有的空文件夹
	emptyDirList := getEmptyDir(filePath)
	for _, row := range emptyDirList {
		err := os.RemoveAll(row)
		if err != nil {
			return err2.ENewFileError(err)
		}
	}

	return nil
}

/*
 * @param dir 文件夹地址
 * @return 所有空文件夹数组
 * @description 获取文件夹下面所有的空文件夹数组
 */
func GetAllEmptyDir(dir string) (*err2.FileError, []string) {
	err, is, _ := IsDirExist(dir)
	if err != nil {
		return err2.ENewFileError(err), nil
	}
	if !is {
		return err2.NewFileError("非文件夹"), nil
	}
	return nil, getEmptyDir(dir)
}

func getEmptyDir(dir string) []string {
	emptyDirList := make([]string, 0)
	info, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, f := range info {
		if f.IsDir() {
			list := getFiles(fmt.Sprint(dir, "\\", f.Name()))
			if list == nil || len(list) == 0 {
				emptyDirList = append(emptyDirList, dir)
			}
		}
	}

	return emptyDirList
}

func getFiles(filePath string) []*os.File {
	fileList := make([]*os.File, 0)
	info, err := ioutil.ReadDir(filePath)
	if err != nil {
		return nil
	}

	for _, f := range info {
		if f.IsDir() {
			list := getFiles(fmt.Sprint(filePath, "\\", f.Name()))
			if list != nil {
				for _, row := range list {
					fileList = append(fileList, row)
				}
			}
		}
	}

	return fileList
}
