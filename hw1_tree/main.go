package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type FileInfoType []os.FileInfo

func (a FileInfoType) Len() int {
	return len(a)
}

func (a FileInfoType) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a FileInfoType) Less(i, j int) bool {
	return a[i].Name() < a[j].Name()
}

func getFilesInfo(path string) (FileInfoType, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("[getFilesInfo]: Error open directory")
	}
	defer file.Close()

	fileInfo, err := file.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("[getFilesInfo]: Error read directory")
	}

	return fileInfo, nil
}

func getFilesForPrint(filesInfo FileInfoType, printFiles bool) FileInfoType {
	if printFiles {
		return filesInfo
	}
	var resultFileInfo = make(FileInfoType, 0, cap(filesInfo))
	for _, file := range filesInfo {
		if file.IsDir() {
			resultFileInfo = append(resultFileInfo, file)
		}
	}

	return resultFileInfo
}

func getSize(file os.FileInfo) string {
	if file.Size() == 0 {
		return "empty"
	}

	return fmt.Sprint(file.Size()) + "b"
}

func printDir(output io.Writer, result string, fileName string, isLastFile bool) {
	if isLastFile {
		fmt.Fprintf(output, result+"└───%s\n", fileName)
		return
	}
	fmt.Fprintf(output, result+"├───%s\n", fileName)
}

func printFile(output io.Writer, result string, file os.FileInfo, isLastFile bool) {
	size := getSize(file)
	if isLastFile {
		fmt.Fprintf(output, result+"└───%s (%s)\n", file.Name(), size)
		return
	}
	fmt.Fprintf(output, result+"├───%s (%s)\n", file.Name(), size)
}

func getResultTree(output io.Writer, path string, printFiles bool, result string) (err error) {
	filesInfo, err := getFilesInfo(path)
	if err != nil {
		return err
	}
	filesInfo = getFilesForPrint(filesInfo, printFiles)
	if err != nil {
		return err
	}

	sort.Sort(filesInfo)
	indexLastFile := len(filesInfo) - 1

	for indexFile, file := range filesInfo {
		var isLastFile bool = indexLastFile == indexFile

		if file.IsDir() {

			printDir(output, result, file.Name(), isLastFile)

			if isLastFile {
				return getResultTree(output, filepath.Join(path, file.Name()), printFiles, result+"\t")
			}

			err = getResultTree(output, filepath.Join(path, file.Name()), printFiles, result+"│\t")
			if err != nil {
				return err
			}

		} else if printFiles {
			printFile(output, result, file, isLastFile)
		}
	}
	return nil
}

func dirTree(output io.Writer, path string, printFiles bool) (err error) {
	return getResultTree(output, path, printFiles, "")
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
