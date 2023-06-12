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
)

// logFunc that shall be used for loggihng
type logFunc func(msg string, logLevel string)

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
	SetLangSuffix(defaulLangSuffix)
	SetLangDir(defaultLangDir)
	SetLang(defaultLang)
	texts = make(map[string]map[string]string)
	texts[defaultLang] = make(map[string]string)

	// load translations
	for _, filename := range getLangFileList() {
		logFunction(fmt.Sprintf(`reading language file '%s'`, filename), `info`)

		var lang = strings.TrimSuffix(filename, langSuffix)
		langFound = append(langFound, lang)
		filename = langDir + filename
		_, err := os.Stat(filename)
		if err != nil {
			logFunction(fmt.Sprintf(`translation file '%s' could not be accessed. File and rights ok?`, filename), LogLevelError)
			continue
		}

		// read file
		jsonData, err := os.ReadFile(filename) // the file is inside the local directory
		if err != nil {
			logFunction(fmt.Sprintf(`translation file '%s' could not be read. File and rights ok?`, filename), LogLevelError)
			continue
		}

		// parse file
		langData := &jsonLanguageData{}
		if err := json.Unmarshal(jsonData, langData); err != nil {
			logFunction(fmt.Sprintf(`translation for language '%s' has an invalid file format. JSON structure ok?`, lang), LogLevelError)
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
}

// SetLangDir from which the files shall be read
func SetLangDir(dir string) {
	langDir = dir + `/`
}

// Set language
func SetLang(lang string) error {
	curLang = lang

	if lang == `` {
		return errors.New(`i18n SetLang: cannot use empty language`)
	}

	// language file exists?
	for _, filename := range getLangFileList() {
		var langFound = strings.TrimSuffix(filename, langSuffix)
		if lang == langFound {
			return nil
		}
	}

	logFunction(
		fmt.Sprintf("language '%s' does not have any langauge file. Will try to fall back to I18N_DEFAULT_LANG\n", lang),
		LogLevelWarn,
	)

	// now try to fallback to ENV defined language
	curLang = os.Getenv(`I18N_DEFAULT_LANG`)
	if curLang == `` {
		logFunction(
			fmt.Sprintf("I18N_DEFAULT_LANG is also undefined. Running out of options. Now using default language '%s' \n", defaultLang),
			LogLevelError,
		)
		curLang = defaultLang // still unsuccessful?
	}

	return nil
}

// overwrite log function
func SetLogFunc(f logFunc) {
	logFunction = f
}

// Set language
func SetLangSuffix(suffix string) {
	langSuffix = suffix
}

func Get(id string) string {
	if curLang == `` {
		return `No language defined. Set language first`
	}

	text, exists := texts[curLang][id]
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

// IsLangFileConsistencyOk does consistency checks on language files
func IsLangFileConsistencyOk() bool {
	if curLang == `` {
		curLang = defaultLang
	}

	logMsg(fmt.Sprintf(`Using language '%s' as reference...`, curLang), LogLevelInfo)

	_, exists := texts[curLang]
	if !exists {
		logMsg(fmt.Sprintf(`Reference language file for language '%s' does not exists. Please name another language`, curLang), LogLevelError)
		return false
	}

	logMsg(fmt.Sprintf(`Languages found: %s. Checking consistency...`, strings.Join(langFound, `, `)), LogLevelInfo)

	referenceLangLength := len(texts[curLang])

	ret := true
	for lang, langtext := range texts {
		if lang == curLang {
			continue
		}

		// check count
		if referenceLangLength != len(langtext) {
			logMsg(fmt.Sprintf(`Language file '%s' (%d entries) differes from the reference language '%s' (%d entries)`, lang, len(langtext), curLang, referenceLangLength), LogLevelWarn)
			ret = false
		}

		// check, if elements that exists in reference language file does also exists in the others
		for key := range texts[curLang] {
			_, textExists := langtext[key]
			if !textExists {
				logMsg(fmt.Sprintf(`Language key '%s' was not found in language '%s'`, key, lang), LogLevelWarn)
			}
		}

		// now the other way around: check if there are additional texts in other language files that are not included in the reference file
		for key := range langtext {
			_, textExists := texts[curLang][key]
			if !textExists {
				logMsg(fmt.Sprintf(`Language key '%s' in language '%s' does not exists in the reference language '%s'`, key, lang, curLang), LogLevelWarn)
			}
		}
	}

	return ret
}
