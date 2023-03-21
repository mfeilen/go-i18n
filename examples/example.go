package examples

import (
	"fmt"

	"github.com/mfeilen/go-i18n"
)

func run() {
	// init i18n stuff
	i18n.SetLangDir(`./lang`)        // default
	i18n.SetLang(`en`)               // default
	i18n.SetLangSuffix(`.json`)      // default, lang filename is as [somelang].json
	i18n.SetLogFunc(mySimpleLogFunc) // default uses https://pkg.go.dev/log

	// output data
	fmt.Println(i18n.Get(`my.text`))
}

func mySimpleLogFunc(msg string, logLevel string) {
	// do some logging
}
