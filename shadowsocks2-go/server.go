package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"yamux"
)

var (
	cport string
	tport string
)
var session *yamux.Session

func init() {
	flag.StringVar(&tport, "t", "9902", "The port of the target connection,Default 9902")
	flag.StringVar(&cport, "c", "127.0.0.1:9901", "What port do you connect to,Default 127.0.0.1:9901")
}
func main() {
	flag.Usage = func() {
		fmt.Println("server -c 9902 -t 127.0.0.1:9901")
	}
	flag.Parse()
	go targetPort(tport)
	yourAddr(cport)
}
func targetPort(address string) (err error) {
	log.Println("Waiting for target in port: " + address)
	ln, err := net.Listen("tcp", "0.0.0.0:"+address)
	if err != nil {
		return errors.New("start port error: " + err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return errors.New("Accept port error")
		}
		log.Println("Connected to target: ", conn.RemoteAddr().String())
		session, err = yamux.Client(conn, nil)
	}
}
func yourAddr(address string) error {
	log.Println("Waiting for clients in port: " + address)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return errors.New("start port error: " + err.Error())
	}
	for {
		conn, err := ln.Accept()
		log.Println("Connected to client: ", conn.RemoteAddr().String())
		if err != nil {
			return errors.New("Accept port error")
		}
		if session == nil {
			log.Println("Error target not connect")
			conn.Close()
			continue
		}
		stream, err := session.Open()
		if err != nil {
			return err
		}
		go func() {
			io.Copy(conn, stream)
			conn.Close()
		}()
		go func() {
			io.Copy(stream, conn)
			stream.Close()
		}()
	}
}
