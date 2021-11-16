package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

func CheckProxyAlive(proxy string) (bool,string) {
	var resp = make([]byte, 128)
	conn, err := net.DialTimeout("tcp", proxy,time.Second*3)
	if err != nil {
		return false, "connect error"
	}
	defer conn.Close()
	conn.Write([]byte{0x05, 0x01,0x00})
	conn.SetReadDeadline(time.Now().Add(time.Second*3))
	n, _ := io.ReadFull(conn, resp[:2])
	if n != 2 || (int(resp[0])!=5 || int(resp[1])!=0) {
		return false, "USERNAME/PASSWORD"
	}
	conn.Write([]byte{0x05,0x01,0x00,0x01,0x7a,0x33,0xa2,0xf9,0x00,0x15})
	n, _ = io.ReadFull(conn, resp[2:4])
	if n != 2 || (int(resp[0])!=5 || int(resp[1])!=0) {
		return false, "target orrer"
	}
	n, err = io.ReadFull(conn, resp[4:14])
	if n!=10{
		return false,"空"
	}
	conn.Read(resp)
	fmt.Println("[+] 代理地址："+proxy+" 外网地址："+string(resp[3:17]))
	return true,"1"
}


func main(){
	var wg sync.WaitGroup
	f, _ := os.Open("1.txt")
	defer f.Close()
	scanner:=bufio.NewScanner(f)
	flag:=true
	for flag{
		for i:=0;i<30;i++ {
			if !scanner.Scan(){flag =false;break}
			line := scanner.Text()
			wg.Add(1)
			go func() {
				CheckProxyAlive(line)
				defer wg.Done()
			}()
		}
		wg.Wait()
	}
}
