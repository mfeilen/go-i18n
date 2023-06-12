package main

import (
	"fmt"

	"github.com/mfeilen/go-i18n"
)

func main() {
	// init i18n stuff - full set
	i18n.SetLangDir(`./lang`)        // default
	i18n.SetLang(`en`)               // default
	i18n.SetLangSuffix(`.json`)      // default, lang filename is as [somelang].json
	i18n.SetLogFunc(mySimpleLogFunc) // default uses https://pkg.go.dev/log

	if !i18n.IsLangFileConsistencyOk() {
		fmt.Println(`Language files are not consistent. See log for more information`)
	}

	// output data
	fmt.Printf("\nValid language text is: %s\n", i18n.Get(`module.function.title`))
	fmt.Println(fmt.Sprintf(`Invalid language text is: %s`, i18n.Get(`some.invalid.language.text`)))
}

func mySimpleLogFunc(msg string, logLevel string) {
	// do some logging
}
