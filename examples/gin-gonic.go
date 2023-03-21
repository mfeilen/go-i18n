package examples

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mfeilen/go-i18n"
)

func main() {

	// init i18n stuff
	i18n.SetLangDir(`./lang`)   // default
	i18n.SetLang(`en`)          // default
	i18n.SetLangSuffix(`.json`) // default, lang filename is as [somelang].json
	i18n.SetLogFunc(myLogFunc)  // default uses https://pkg.go.dev/log

	// set router and and
	router := gin.Default()
	setMiddleware(router)
	// set Gin gonic routes and start server - see gin documentation

	// Usage in some Gin context / HandlerFunc: fmt.Println(i18n.Get(`module.function.title`))
}

// setMiddleware of the given router
func setMiddleware(router *gin.Engine) {

	// Middleware stuff
	router.Use(
		gin.Recovery(),       // recovers from any panics and writes a 500 if there was one.
		setLangFromBrowser(), // detects the browser language and sets it
	)
}

// setLangFromBrowser handlerFunc will be triggered on each request
func setLangFromBrowser() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Request.Header[`Accept-Language`]
		if len(lang) == 0 {
			c.Next()
			return
		}

		langStr := strings.Split(lang[0], `;`)
		if len(langStr) == 0 {
			c.Next()
			return
		}
		// fav lang list
		langList := strings.Split(langStr[0], `,`)
		i18n.SetLang(langList[0]) // can be en|en-us|... json-file should be named accordingly
		c.Next()
	}
}

func myLogFunc(msg string, logLevel string) {
	// do some logging
}
