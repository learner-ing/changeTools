package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const timeout = 5

func main() {
	args := os.Args
	if len(os.Args) <= 1 {
		os.Exit(0)
	}
	port := checkPort(args[1])
	port2host(port)
}
func checkPort(port string) string {
	PortNum, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalln("[x]", "port should be a number")
	}
	if PortNum < 1 || PortNum > 65535 {
		log.Fatalln("[x]", "port should be a number and the range is [1,65536)")
	}
	return port
}
func getAddrs()[]string{
	f, err := os.Open("config.ini")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	addrs:=[]string{}
	for scanner.Scan() {
		addrs = append(addrs, scanner.Text())
	}
	return addrs
}

func port2host(allowPort string) {
	server := start_server("127.0.0.1:" + allowPort)
	targetAddresss:=getAddrs()
	targetAddress:=""
	i:=0
	for {
		conn := accept(server)
		if conn == nil {
			continue
		}
		targetAddress=targetAddresss[i%len(targetAddresss)]
		go func(targetAddress string) {
			log.Println("[+]", "start connect host:["+targetAddress+"]")
			target, err := net.Dial("tcp", targetAddress)
			if err != nil {
				log.Println("[x]", "connect target address ["+targetAddress+"] faild. retry in ", timeout, "seconds. ")
				conn.Close()
				log.Println("[←]", "close the connect at local:["+conn.LocalAddr().String()+"] and remote:["+conn.RemoteAddr().String()+"]")
				time.Sleep(timeout * time.Second)
				return
			}
			log.Println("[→]", "connect target address ["+targetAddress+"] success.")
			forward(target, conn)
		}(targetAddress)
		i++
	}
}
func start_server(address string) net.Listener {
	log.Println("[+]", "try to start server on:["+address+"]")
	server, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("[x]", "listen address ["+address+"] faild.")
	}
	log.Println("[√]", "start listen at address:["+address+"]")
	return server
}

func accept(listener net.Listener) net.Conn {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("[x]", "accept connect ["+conn.RemoteAddr().String()+"] faild.", err.Error())
		return nil
	}
	log.Println("[√]", "accept a new client. remote address:["+conn.RemoteAddr().String()+"], local address:["+conn.LocalAddr().String()+"]")
	return conn
}

func forward(conn1 net.Conn, conn2 net.Conn) {
	log.Printf("[+] start transmit. [%s],[%s] <-> [%s],[%s] \n", conn1.LocalAddr().String(), conn1.RemoteAddr().String(), conn2.LocalAddr().String(), conn2.RemoteAddr().String())
	var wg sync.WaitGroup
	// wait tow goroutines
	wg.Add(2)
	go connCopy(conn1, conn2, &wg)
	go connCopy(conn2, conn1, &wg)
	//blocking when the wg is locked
	wg.Wait()
}

func connCopy(conn1 net.Conn, conn2 net.Conn, wg *sync.WaitGroup) {
	logFile := openLog(conn1.LocalAddr().String(), conn1.RemoteAddr().String(), conn2.LocalAddr().String(), conn2.RemoteAddr().String())
	if logFile != nil {
		w := io.MultiWriter(conn1, logFile)
		io.Copy(w, conn2)
	} else {
		io.Copy(conn1, conn2)
	}
	conn1.Close()
	log.Println("[←]", "close the connect at local:["+conn1.LocalAddr().String()+"] and remote:["+conn1.RemoteAddr().String()+"]")
	wg.Done()
}
func openLog(address1, address2, address3, address4 string) *os.File {
	args := os.Args
	argc := len(os.Args)
	var logFileError error
	var logFile *os.File
	if argc > 5 && args[4] == "-log" {
		address1 = strings.Replace(address1, ":", "_", -1)
		address2 = strings.Replace(address2, ":", "_", -1)
		address3 = strings.Replace(address3, ":", "_", -1)
		address4 = strings.Replace(address4, ":", "_", -1)
		timeStr := time.Now().Format("2006_01_02_15_04_05") // "2006-01-02 15:04:05"
		logPath := args[5] + "/" + timeStr + args[1] + "-" + address1 + "_" + address2 + "-" + address3 + "_" + address4 + ".log"
		logPath = strings.Replace(logPath, `\`, "/", -1)
		logPath = strings.Replace(logPath, "//", "/", -1)
		logFile, logFileError = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE, 0666)
		if logFileError != nil {
			log.Fatalln("[x]", "log file path error.", logFileError.Error())
		}
		log.Println("[√]", "open test log file success. path:", logPath)
	}
	return logFile
}