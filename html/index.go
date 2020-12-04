package html

//go:generate sh -c "(echo 'package html' && echo && printf 'const indexFile = `' && cat index.html && echo '`') > file.go"

func Index() string {
	return indexFile
}
