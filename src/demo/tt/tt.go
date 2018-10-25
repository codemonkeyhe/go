// tt project tt.go
package tt

import "strings"
import "os"

func TestP() {

}

func GetFileName() string {
	filename := os.Args[0]
	pos := strings.LastIndex(filename, "/")
	if pos > 0 {
		filename = filename[pos+1:]
	}
	pos = strings.LastIndex(filename, "\\")
	if pos > 0 {
		filename = filename[pos+1:]
	}
	return filename
}

func Help() {

}
