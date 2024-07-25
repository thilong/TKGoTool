package webdav_uploader

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/studio-b12/gowebdav"
	"github.com/tidwall/gjson"
)

func PrintUsage() {
	fmt.Println(`Use -W {path} [-Ws {sub mode}] to upload a folder.
When sub mode is not set, uploader will use .tk.webdav as configuration,
otherwise, .tk.webdav.{sub mode} will be used.
	`)
}

type WebDavConfig struct {
	Server   string
	RootPath string
	Uid      string
	Pwd      string
	Exclude  []string
	Reboot   string
}

type WebdavUploadResult struct {
	result map[string]int64
	error  interface{}
}

func (result *WebdavUploadResult) Save(path string) {

	fileHandler, openError := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if openError != nil {
		fmt.Println("[Webdav] save record to", path, "failed.", openError)
		return
	}
	defer fileHandler.Close()
	for file, mtime := range result.result {
		fileHandler.WriteString(fmt.Sprintf("%s | %d\n", file, mtime))
	}
	fileHandler.WriteString("EOF")
}

func (result *WebdavUploadResult) Parse(path string) {
	fileHandler, openError := os.OpenFile(path, os.O_RDONLY, 0644)
	if openError != nil {
		fmt.Println("[Webdav] Can't open records, upload all.")
		return
	}
	defer fileHandler.Close()
	bufReader := bufio.NewScanner(fileHandler)
	for bufReader.Scan() {
		lineText := bufReader.Text()
		if lineText == "EOF" {
			break
		}
		mapComs := strings.Split(lineText, " | ")
		if len(mapComs) > 1 {
			if parsedTime, parsedError := strconv.ParseInt(mapComs[1], 10, 64); parsedError == nil {
				result.result[mapComs[0]] = parsedTime
			}
		}
	}
}

func parseConfigFile(file string) *WebDavConfig {

	config := &WebDavConfig{}

	fileBytes, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error read config file ", file)
	}
	fileStr := string(fileBytes)
	config.Server = gjson.Get(fileStr, "Server").String()
	config.RootPath = gjson.Get(fileStr, "RootPath").String()
	config.Uid = gjson.Get(fileStr, "Uid").String()
	config.Pwd = gjson.Get(fileStr, "Pwd").String()
	excludeResults := gjson.Get(fileStr, "Exclude").Array()
	config.Exclude = []string{}
	for _, eResult := range excludeResults {
		config.Exclude = append(config.Exclude, eResult.String())
	}
	config.Reboot = gjson.Get(fileStr, "Reboot").String()

	return config
}

func Upload(folder string, sub_mode string) {
	var configFile string = ".tk.webdav"
	if len(sub_mode) > 0 {
		configFile = configFile + "." + sub_mode
	}
	configFile = path.Join(folder, configFile)

	if _, err := os.Stat(configFile); err != nil {
		fmt.Println("Can't find the config file \"", configFile, "\"")
		PrintUsage()
		return
	}
	config := parseConfigFile(configFile)
	//fmt.Printf("%+v\n", config)

	client := gowebdav.NewClient(config.Server, config.Uid, config.Pwd)

	if err := client.Connect(); err != nil {
		fmt.Println("[Webdav] Can't connect to server, ", err)
		return
	} else {
		fmt.Println("[Webdav] connected ...")
	}

	result := &WebdavUploadResult{result: make(map[string]int64)}
	result.Parse(configFile + ".result")

	uploadFolder(folder, config.RootPath, config, client, result)
	result.Save(configFile + ".result")
	if result.error != nil {
		fmt.Println("[Webdav] error:", result.error)
	} else {
		if len(config.Reboot) > 0 {
			response, resErr := http.Get(config.Reboot)
			if resErr == nil && response.StatusCode == 200 {
				fmt.Println("[Webdav] reboot ...")
			} else {
				fmt.Println("[Webdav] reboot failed.")
			}
		}
		fmt.Println("[Webdav] done.")
	}

}

func uploadFolder(folder string, toFolder string, config *WebDavConfig, client *gowebdav.Client, result *WebdavUploadResult) {
	//fmt.Println("[Webdav]", folder, "...")
	dirEntries, err := os.ReadDir(folder)
	if err != nil {
		result.error = err
		fmt.Println("[Webdav] Can't read files in ", folder)
		return
	}
	for _, entry := range dirEntries {
		fileName := entry.Name()
		if strings.HasPrefix(fileName, ".tk.webdav") {
			continue
		}
		fromFile := path.Join(folder, fileName)
		toFile := path.Join(toFolder, fileName)
		//if path is dir, then do upload folder
		if entry.IsDir() {
			mkdirErr := client.MkdirAll(toFile, 0644)
			if mkdirErr != nil {
				fmt.Println("[Webdav] can't create folder ,", mkdirErr)
			}
			uploadFolder(fromFile, toFile, config, client, result)
			continue
		}
		excluded := false
		//check config for exclude
		for _, pattern := range config.Exclude {
			if found, _ := regexp.MatchString(pattern, fileName); found {
				excluded = true
				break
			}
		}
		if excluded {
			fmt.Println("[Webdav]", path.Join(folder, entry.Name()), "excluded ...")
			continue
		}

		//check file info, get modify time
		fileInfo, err := os.Stat(fromFile)
		if err != nil {
			fmt.Println("[Webdav]", fromFile, "-> error,", err)
			result.error = err
			break
		}
		recordTime := result.result[fromFile]
		if recordTime == fileInfo.ModTime().Unix() {
			//fmt.Println("[Webdav]", path.Join(folder, entry.Name()), "->", path.Join(toFolder, entry.Name()), "ignored, file is not modified.")
			continue
		}
		//do upload
		fileHandler, openErr := os.Open(fromFile)
		if openErr != nil {
			result.error = openErr
			fmt.Println("[Webdav]", path.Join(folder, entry.Name()), "->", path.Join(toFolder, entry.Name()), "failed,", openErr)
			break
		}

		if writeErr := client.WriteStream(toFile, fileHandler, 0644); writeErr != nil {
			fmt.Println("[Webdav]", "->", path.Join(toFolder, entry.Name()), "failed,", writeErr)
		} else {
			fmt.Println("[Webdav]", "->", path.Join(toFolder, entry.Name()), "ok ...")
		}
		//fmt.Println("[Webdav]", path.Join(folder, entry.Name()), "->", path.Join(toFolder, entry.Name()), "ok ...")
		fileHandler.Close()
		result.result[fromFile] = fileInfo.ModTime().Unix()
	}
}
