package logger

import (
	"fmt"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/bncrypted/apidor/pkg/http"
)

// Flags is a logger struct that holds command line flags for customising the logging output
type Flags struct {
	DefinitionFile string
	LocalCertFile  string
	LogFile        string
	ProxyURI       string
	Rate           int
	IsDebug        bool
}

var isDebug bool
var isWriteToFile bool
var logfile *os.File

// Init is a logger function that initialises the logger based on the given flags
func Init(flags Flags) {
	initFile(flags.LogFile)
	initDebug(flags.IsDebug)
}

func initFile(filename string) {
	if filename == "" {
		isWriteToFile = false
		return
	}

	isWriteToFile = true
	var err error
	logfile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		isWriteToFile = false
		Error("Could not open logfile, continuing without writing to file")
	}
}

func initDebug(debug bool) {
	isDebug = debug
}

// Close is a logger function that closes the log file
func Close() {
	if isWriteToFile {
		err := logfile.Close()
		if err != nil {
			Error("Could not close logfile")
		}
	}
}

// Logo is a logger function that prints the apidor logo
func Logo() {
	logo := `

	 █████╗ ██████╗ ██╗██████╗  ██████╗ ██████╗ 
	██╔══██╗██╔══██╗██║██╔══██╗██╔═══██╗██╔══██╗
	███████║██████╔╝██║██║  ██║██║   ██║██████╔╝
	██╔══██║██╔═══╝ ██║██║  ██║██║   ██║██╔══██╗
	██║  ██║██║     ██║██████╔╝╚██████╔╝██║  ██║
	╚═╝  ╚═╝╚═╝     ╚═╝╚═════╝  ╚═════╝ ╚═╝  ╚═╝
												
	`

	writeln(logo)
}

// RunInfo is a logger function that prints some base information about the current execution
func RunInfo(baseURI string, endpointsCount int, flags Flags) {
	writeln("API: " + baseURI)
	writeln("Endpoints: " + strconv.Itoa(endpointsCount))
	writeln("Time: " + time.Now().String())
	writeln("")

	if flags.LocalCertFile != "" {
		writeln("Cert: " + flags.LocalCertFile)
	}
	if flags.DefinitionFile != "" {
		writeln("Definition: " + flags.DefinitionFile)
	}
	if flags.LogFile != "" {
		writeln("Log: " + flags.LogFile)
	}
	if flags.ProxyURI != "" {
		writeln("Proxy: " + flags.ProxyURI)
	}
	writeln("Rate: " + strconv.Itoa(flags.Rate) + " req/s")
	if flags.IsDebug {
		writeln("Debugging: on")
	}

	writeln("")
}

// Starting is a logger function that prints a start message
func Starting() {
	writeln("Starting...")
	writeln("")
}

// Finished is a logger function that prints a finished message
func Finished() {
	if !isDebug {
		writeln("")
	}
	writeln("Done, nice one 👊")
	writeln("")
}

// TestPrefix is a logger function that prints the endpoint and the test name that is taking place
func TestPrefix(requestID int, endpoint string, testName string) {
	prefix := "[" + strconv.Itoa(requestID) + "][" + endpoint + "][" + testName + "] "
	if isDebug {
		writeln(prefix)
		writeln("")
	} else {
		write(prefix)
	}
}

// TestResult is a logger function that prints the result of the test
func TestResult(result string) {
	if isDebug {
		writeln(result)
		writeln("")
	} else {
		writeln(result)
	}
}

// Message is a logger function that prints a given message
func Message(message string) {
	writeln(message)
}

// DebugMessage is a logger function that prints if the debug flag is set
func DebugMessage(message string) {
	if isDebug {
		Message(message)
		writeln("")
	}
}

// Error is a logger function that prints an error message
func Error(message string) {
	writeln("Error: " + message)
}

// DebugError is a logger function that prints an error message if the debug flag is set
func DebugError(message string) {
	if isDebug {
		Error(message)
		writeln("")
	}
}

// Fatal is a logger function that prints a fatal error message
func Fatal(message string) {
	writeln("")
	writeln("Fatal: " + message)
	writeln("Exiting")
	writeln("")
}

// DumpRequest is a logger function that prints a request dump
func DumpRequest(requestOptions http.RequestOptions) {
	if !isDebug {
		writeln("")
		defer writeln("")

		request, err := http.CreateRequest(requestOptions)
		if err != nil {
			writeln("Error dumping request: " + err.Error())
			return
		}
		requestDump, err := httputil.DumpRequest(request, true)
		if err != nil {
			writeln("Error dumping request: " + err.Error())
			return
		}

		writeln(string(requestDump))
	}
}

func write(message string) {
	fmt.Print(message)
	if isWriteToFile {
		logfile.WriteString(message)
	}
}

func writeln(message string) {
	fmt.Println(message)
	if isWriteToFile {
		logfile.WriteString(message + "\n")
	}
}
