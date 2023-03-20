package examples

import (
	"strings"

	"github.com/mfeilen/go-i18n"

	"github.com/gin-gonic/gin"
)

func main() {

	// init i18n stuff
	i18n.SetLangDir(`./lang`)   // default
	i18n.SetLang(`en`)          // default
	i18n.SetLangSuffix(`.json`) // default, lang filename is as [somelang].json
	i18n.SetLogFunc(myLog)      // default uses https://pkg.go.dev/log

	// set router and and
	router := gin.Default()
	setMiddleware(router)
	// set Gin gonic routes and start server - see gin documentation

	// Usage in some Gin context / HandlerFunc: fmt.Println(i18n.Get(`my.text`))
}

// setMiddleware of the given router
func setMiddleware(router *gin.Engine) {

	// Middleware stuff
	router.Use(
		gin.Recovery(),       // recovers from any panics and writes a 500 if there was one.
		setI18nFromBrowser(), // detects the browser language and sets it
	)
}

// setI18nFromBrowser handlerFunc will be triggered on each request
func setI18nFromBrowser() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.Header[`Accept-Language`]
		if len(lang) > 0 {
			langStr := strings.Split(lang[0], `;`)
			if len(langStr) > 0 {
				// fav lang list
				langList := strings.Split(langStr[0], `,`)

				// get the first one
				firstLang := langList[0][0:2] // get the first 2 letters of the first language
				i18n.SetLang(firstLang)
			}
		}
		c.Next()
	}
}

func myLog(msg string, logLevel string) {
	// do some logging
}
