package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	// defaults
	defaultLang      = `en`
	defaultLangDir   = `./lang`
	defaulLangSuffix = `.json`
	LogLevelError    = `error`
	LogLevelWarn     = `warn`
	LogLevelInfo     = `info`
)

var (
	// curLang current language
	curLang string
	// texts loaded for all languages as map[language]map[key]language-text
	texts map[string]map[string]string
	// languages found
	langFound []string
	// logFunction used to log errors / infos. Can be overwritten with setLogFunc
	logFunction logFunc
	// langDir ectory that contains the language files
	langDir string
	// the expected file suffix for the language files
	langSuffix string
	// readfile func - in case it should be overwritten
	readFileFunction readFileFunc
)

// logFunc that shall be used for loggihng
type logFunc func(msg string, logLevel string)
type readFileFunc func(name string) ([]byte, error)

// jsonLanguageData contains the parsed data
type jsonLanguageData struct {
	Rows []jsonLanguageDataElement `json:"lang"`
}

type jsonLanguageDataElement struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}

// init configration and load language files
func init() {

	// set defaults
	SetLogFunc(logMsg)
	SetReadFileFunc(readFile)
	SetLangSuffix(defaulLangSuffix)
	SetLangDir(defaultLangDir)
	texts = make(map[string]map[string]string)

	// load translations
	for _, filename := range getLangFileList() {
		logFunction(fmt.Sprintf(`reading language file '%s'`, filename), `info`)

		var lang = strings.TrimSuffix(filename, langSuffix)
		langFound = append(langFound, lang)
		filename = langDir + filename
		_, err := os.Stat(filename)
		if err != nil {
			logFunction(fmt.Sprintf(`i18n: translation file '%s' could not be accessed. File and rights ok?`, filename), LogLevelError)
			continue
		}

		// read file
		jsonData, err := readFileFunction(filename) // the file is inside the local directory
		if err != nil {
			logFunction(fmt.Sprintf(`i18n: translation file '%s' could not be read. File and rights ok?`, filename), LogLevelError)
			continue
		}

		// parse file
		langData := &jsonLanguageData{}
		if err := json.Unmarshal(jsonData, langData); err != nil {
			logFunction(fmt.Sprintf(`i18n: translation for language '%s' has an invalid file format. JSON structure ok?`, lang), LogLevelError)
			continue
		}

		_, exists := texts[lang]
		if !exists {
			texts[lang] = make(map[string]string)
		}

		// set language values
		for _, row := range langData.Rows {
			texts[lang][row.Id] = row.Text
		}
	}

	// choose initial language after all language files were loaded
	curLang = resolveDefaultLang()
}

// SetLangDir from which the files shall be read
func SetLangDir(dir string) {
	if strings.TrimSpace(dir) == `` {
		langDir = defaultLangDir + `/`
		logFunction(
			fmt.Sprintf("i18n: empty language directory provided. Falling back to default '%s'", langDir),
			LogLevelWarn,
		)
		return
	}

	langDir = strings.TrimRight(dir, `/\`) + `/`
}

// SetLang that shall be usesd if no language is given in Get()
func SetLang(lang string) error {
	lang = strings.TrimSpace(lang)
	if lang == `` {
		return errors.New(`i18n: cannot set empty language`)
	}

	// check if the language file exists, then load it
	if _, exists := texts[lang]; exists {
		curLang = lang
		return nil
	}

	logFunction(
		fmt.Sprintf("i18n: language '%s' is not loaded. Falling back to default language", lang),
		LogLevelWarn,
	)

	curLang = resolveDefaultLang()
	return nil
}

func resolveDefaultLang() string {

	// fallback to the default language configured in the env file
	envLang := strings.TrimSpace(os.Getenv(`I18N_DEFAULT_LANG`))
	if envLang != `` {
		if _, exists := texts[envLang]; exists {
			return envLang
		}
		logFunction(
			fmt.Sprintf("i18n: I18N_DEFAULT_LANG '%s' is not loaded. Trying fallback '%s'", envLang, defaultLang),
			LogLevelWarn,
		)
	}

	// fallback to default language (hardcoded in package)
	if _, exists := texts[defaultLang]; exists {
		return defaultLang
	}

	// fallback to first loaded language to keep the package usable
	for _, lang := range langFound {
		if _, exists := texts[lang]; exists {
			logFunction(
				fmt.Sprintf("i18n: default language '%s' is not loaded. Using first loaded language '%s'", defaultLang, lang),
				LogLevelWarn,
			)
			return lang
		}
	}

	logFunction(
		fmt.Sprintf("i18n: no language files loaded. Using fallback language '%s'", defaultLang),
		LogLevelWarn,
	)
	return defaultLang
}

// overwrite log function
func SetLogFunc(f logFunc) {
	logFunction = f
}

func SetReadFileFunc(f readFileFunc) {
	readFileFunction = f
}

// Set language
func SetLangSuffix(suffix string) {
	langSuffix = suffix
}

func Get(id string, lang ...string) string {
	selectedLang := curLang
	if len(lang) > 0 && lang[0] != `` {
		selectedLang = lang[0]
	}

	if selectedLang == `` {
		return `i18n: No language defined. Set language first`
	}

	text, exists := texts[selectedLang][id]
	if !exists {
		return id
	}

	return text
}

// getLangFileList returns the language files w/o pathes
func getLangFileList() []string {
	var fileList []string
	files, err := os.ReadDir(langDir)
	if err != nil {
		logFunction(err.Error(), LogLevelError)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// only .json files
		var extension = filepath.Ext(file.Name())
		if extension != langSuffix {
			logFunction(fmt.Sprintf(`language file '%s' has not the expected suffix '%s'. Skipping it`, file.Name(), langSuffix), LogLevelInfo)
			continue
		}

		fileList = append(fileList, file.Name())
	}

	return fileList
}

// logMsg using golang log package
func logMsg(msg string, logLevel string) {

	switch logLevel {
	case LogLevelError:
		log.Println(`i18n error: ` + msg)
	case LogLevelWarn:
		log.Println(`i18n warn: ` + msg)
	default:
		log.Println(`i18n info: ` + msg)
	}
}

func readFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// IsLangFileConsistencyOk does consistency checks on language files
func IsLangFileConsistencyOk() bool {
	if curLang == `` {
		curLang = defaultLang
	}

	logFunction(fmt.Sprintf(`Using language '%s' as reference...`, curLang), LogLevelInfo)

	_, exists := texts[curLang]
	if !exists {
		logFunction(fmt.Sprintf(`Reference language file for language '%s' does not exists. Please name another language`, curLang), LogLevelError)
		return false
	}

	logFunction(fmt.Sprintf(`Languages found: %s. Checking consistency...`, strings.Join(langFound, `, `)), LogLevelInfo)

	referenceLangLength := len(texts[curLang])

	ret := true
	for lang, langtext := range texts {
		if lang == curLang {
			continue
		}

		// check count
		if referenceLangLength != len(langtext) {
			logFunction(fmt.Sprintf(`Language file '%s' (%d entries) differes from the reference language '%s' (%d entries)`, lang, len(langtext), curLang, referenceLangLength), LogLevelWarn)
			ret = false
		}

		// check, if elements that exists in reference language file does also exists in the others
		for key := range texts[curLang] {
			_, textExists := langtext[key]
			if !textExists {
				logFunction(fmt.Sprintf(`Language key '%s' was not found in language '%s'`, key, lang), LogLevelWarn)
				ret = false
			}
		}

		// now the other way around: check if there are additional texts in other language files that are not included in the reference file
		for key := range langtext {
			_, textExists := texts[curLang][key]
			if !textExists {
				logFunction(fmt.Sprintf(`Language key '%s' in language '%s' does not exists in the reference language '%s'`, key, lang, curLang), LogLevelWarn)
				ret = false
			}
		}
	}

	return ret
}
