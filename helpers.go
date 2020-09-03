package main

import (
	"strings"

	"github.com/karrick/godirwalk"
)

func ReplaceInUUID(id string) string {
	newStr := make([]rune, len(id))
	i := 0
	for _, r := range id {
		switch r {
		case '0':
			newStr[i] = 'q'
		case '1':
			newStr[i] = 'w'
		case '2':
			newStr[i] = 'd'
		case '3':
			newStr[i] = 'e'
		case '4':
			newStr[i] = 'r'
		case '5':
			newStr[i] = 't'
		case '6':
			newStr[i] = 'f'
		case '7':
			newStr[i] = 'g'
		case '8':
			newStr[i] = 'h'
		case '9':
			newStr[i] = 'j'
		default:
			newStr[i] = r
		}
		i++
	}
	return string(newStr)
}

func FindFiles(path, match string, postFix string) (jsonFiles []string) {
	_ = godirwalk.Walk(path, &godirwalk.Options{
		Callback: func(osPathname string, info *godirwalk.Dirent) error {
			if !info.IsDir() {
				if strings.Contains(osPathname, match+postFix) {
					jsonFiles = append(jsonFiles, osPathname)
				}
			}
			return nil
		},
		Unsorted: true,
	})
	return
}
