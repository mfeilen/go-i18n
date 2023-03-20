# I18n Language package
This package is a simple way to use languages in your application.
Create per each language a language file located as defined in I18N_LANG_PATH (exmaple: ./lang/en.json)

## Usage
```
git clone git@github.com:mfeilen/i18n.git
```
Follow exampe as described here: [Default example](examples/example.go)

## Env parameters
| ENV param | usage | example | default |
| --- | --- | --- | --- |
| I18N_DEFAULT_LANG | default language used (ISO 639-1:2002) | de | en |
| I18N_LANG_PATH | files in which the language files are located, no trailing slash | ./myLangFiles | ./lang |

# I18N and Gin Web Framework
You may want to use a browser detection within your Router. Details can be found here [Gin Example](examples/gin-gonic.go)
