package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/long/test/esutil"
	log "github.com/long/test/log"
)

var addr = flag.String("addr", "0.0.0.0", "The address to listen to;")
var Eurl = flag.String("eurl", "http://192.168.1.182:9200/", "The port to listen on; ")
var Picurl = flag.String("picurl", "http://58.118.225.79:41242/", "picture url ")

var port = flag.Int("port", 6000, "The port to listen on; ")
var sec = flag.Int("sec", 10, "the second for query data. ")

var Level = flag.String("level","ErrorLevel","log level")

func main() {
	flag.Parse()

	log.SetLogLevel(*Level)

	fmt.Println(*port)
	src := *addr + ":" + strconv.Itoa(*port)
	listener, err := net.Listen("tcp", src)
	if err != nil {
		log.Log.Errorln(err)
		return
	}
	log.Log.Infof("Listening on %s.\n", src)

	fmt.Println("starting server success.")
	defer listener.Close()

	connArr:=make([]net.Conn,0)

	for {
		conn, err := listener.Accept()//

		connArr = append(connArr,conn)
		if err != nil {
			log.Log.Infoln("some connecion error: ", err)
		}
		go handleConnection(conn,connArr)
	}
}

func handleConnection(conn net.Conn, connArr []net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	log.Log.Infoln("Client connected from ", remoteAddr)

	ech := make(chan error)
	go func(conn net.Conn, ech chan error) {
		buf := make([]byte, 10)
		readMsg, err := conn.Read(buf)
		log.Log.Infoln("Read completed,readMsg:",readMsg,",err:",err)
		ech <- err

	}(conn, ech)

	tick := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-tick.C:
			if !handleMessage(conn, connArr){
				conn.Close()
				return
			}
		case err := <-ech:
			log.Log.Infoln(err, "remoteAddr ", remoteAddr, " close")
			conn.Close()
			return
		}
	}

	log.Log.Infoln("Client at " + remoteAddr + " disconnected.")
}

func handleMessage(conn net.Conn, connArr []net.Conn) bool {
	jsonstring := esutil.PostAction(*sec, *Eurl, *Picurl)
	if jsonstring == nil {
		log.Log.Infoln("the data is nil,remoteArr:",conn.RemoteAddr())
		conn.Write([]byte("\000"))
		return true
	}
	jsonstring = append(jsonstring, []byte("\000")...)
	log.Log.Infoln("jsonstring len: ", len(jsonstring), "\000 data: ", len("\000"))
	_, err := conn.Write(jsonstring)
	if err !=nil{
		log.Log.Infoln("conn.WriteErr:",err)
		return false
	}
	return true
}
