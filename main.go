package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
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
	// optional fs source for embedded files
	fileSystem fs.FS
	// indicates if translations were loaded with current config
	isLoaded bool
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

// init configuration defaults (translations are loaded lazily or via Init).
func init() {

	// set defaults
	SetLogFunc(logMsg)
	SetReadFileFunc(readFile)
	SetLangSuffix(defaulLangSuffix)
	SetLangDir(defaultLangDir)
	texts = make(map[string]map[string]string)
	langFound = nil
	isLoaded = false
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
	isLoaded = false
}

// SetLang that shall be usesd if no language is given in Get()
func SetLang(lang string) error {
	ensureLoaded()

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
	isLoaded = false
}

// Set language
func SetLangSuffix(suffix string) {
	langSuffix = suffix
	isLoaded = false
}

// SetFS sets an alternative file system source for language files (e.g. embed.FS).
func SetFS(fsys fs.FS) {
	fileSystem = fsys
	isLoaded = false
}

// Init triggers loading/reloading language files with the current configuration.
func Init() {
	loadTranslations()
}

func Get(id string, lang ...string) string {
	ensureLoaded()

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
	var files []fs.DirEntry
	var err error

	if fileSystem != nil {
		files, err = fs.ReadDir(fileSystem, getFSDirPath())
	} else {
		files, err = os.ReadDir(langDir)
	}

	if err != nil {
		logFunction(err.Error(), LogLevelError)
		return fileList
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

func ensureLoaded() {
	if isLoaded {
		return
	}
	loadTranslations()
}

func loadTranslations() {
	texts = make(map[string]map[string]string)
	langFound = nil

	for _, filename := range getLangFileList() {
		logFunction(fmt.Sprintf(`reading language file '%s'`, filename), LogLevelInfo)

		lang := strings.TrimSuffix(filename, langSuffix)
		langFound = append(langFound, lang)

		jsonData, err := readLangFile(filename)
		if err != nil {
			logFunction(fmt.Sprintf(`i18n: translation file '%s' could not be read. File and rights ok?`, filename), LogLevelError)
			continue
		}

		langData := &jsonLanguageData{}
		if err := json.Unmarshal(jsonData, langData); err != nil {
			logFunction(fmt.Sprintf(`i18n: translation for language '%s' has an invalid file format. JSON structure ok?`, lang), LogLevelError)
			continue
		}

		if _, exists := texts[lang]; !exists {
			texts[lang] = make(map[string]string)
		}

		for _, row := range langData.Rows {
			texts[lang][row.Id] = row.Text
		}
	}

	curLang = resolveDefaultLang()
	isLoaded = true
}

func readLangFile(filename string) ([]byte, error) {
	if fileSystem != nil {
		return fs.ReadFile(fileSystem, getFSDirPath()+`/`+filename)
	}

	fullPath := langDir + filename
	if _, err := os.Stat(fullPath); err != nil {
		return nil, err
	}

	return readFileFunction(fullPath)
}

func getFSDirPath() string {
	dir := strings.TrimSpace(langDir)
	if dir == `` {
		dir = defaultLangDir
	}

	dir = strings.ReplaceAll(dir, `\`, `/`)
	dir = strings.TrimPrefix(dir, `./`)
	dir = strings.TrimPrefix(dir, `/`)
	dir = strings.TrimSuffix(dir, `/`)
	if dir == `` {
		return `.`
	}

	return path.Clean(dir)
}

// IsLangFileConsistencyOk does consistency checks on language files
func IsLangFileConsistencyOk() bool {
	ensureLoaded()

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
