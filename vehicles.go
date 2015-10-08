package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func getidbcd(id string) []byte {
	idbcd := []byte{01}
	tmpint, _ := strconv.Atoi(id[1:3])
	tmpbcd := tmpint/10*16 + tmpint%10
	tmp := byte(tmpbcd)
	idbcd = append(idbcd, tmp)

	tmpint, _ = strconv.Atoi(id[3:5])
	tmpbcd = tmpint/10*16 + tmpint%10
	tmp = byte(tmpbcd)
	idbcd = append(idbcd, tmp)

	tmpint, _ = strconv.Atoi(id[5:7])
	tmpbcd = tmpint/10*16 + tmpint%10
	tmp = byte(tmpbcd)
	idbcd = append(idbcd, tmp)

	tmpint, _ = strconv.Atoi(id[7:9])
	tmpbcd = tmpint/10*16 + tmpint%10
	tmp = byte(tmpbcd)
	idbcd = append(idbcd, tmp)

	tmpint, _ = strconv.Atoi(id[9:11])
	tmpbcd = tmpint/10*16 + tmpint%10
	tmp = byte(tmpbcd)
	idbcd = append(idbcd, tmp)

	return idbcd

}

func CheckSum(cmd []byte, cmdlen uint16) byte {
	temp := cmd[0]
	for i := uint16(1); i < cmdlen; i++ {
		temp ^= cmd[i]
	}

	return temp
}

func sendauthcmd(id string, conn *net.TCPConn) {
	authcmd := []byte{0x7e, 0x01, 0x02, 0x00, 0x0b}
	authcmd = append(authcmd, getidbcd(id)...)
	authcmd = append(authcmd, 0)
	authcmd = append(authcmd, 1)
	authcmd = append(authcmd, []byte(id)...)
	checksumbyte := authcmd[1:len(authcmd)]
	checksum := CheckSum(checksumbyte, uint16(len(checksumbyte)))
	authcmd = append(authcmd, checksum)
	//authcmd = append(authcmd, 1)
	authcmd = append(authcmd, 0x7e)
	log.Printf("%x\n", authcmd)

	_, err := conn.Write(authcmd)
	if err != nil {
		log.Println(err.Error())
	}
}

func addtime() []byte {
	var cmd []byte
	curtime := time.Now()

	year := curtime.Year() - 2000
	tmp := year/10*16 + year%10
	cmd = append(cmd, byte(tmp))

	month := curtime.Month()
	tmp = int(month)/10*16 + int(month)%10
	cmd = append(cmd, byte(tmp))

	day := curtime.Day()
	tmp = day/10*16 + day%10
	cmd = append(cmd, byte(tmp))

	hour := curtime.Hour()
	tmp = hour/10*16 + hour%10
	cmd = append(cmd, byte(tmp))

	minute := curtime.Minute()
	tmp = minute/10*16 + minute%10
	cmd = append(cmd, byte(tmp))

	second := curtime.Second()
	tmp = second/10*16 + second%10
	cmd = append(cmd, byte(tmp))

	return cmd

}

func sendposcmd(id string, conn *net.TCPConn, wg *sync.WaitGroup) bool {
	poscmd := []byte{0x7e, 0x02, 0x00, 0x00, 0x4c}
	poscmd = append(poscmd, getidbcd(id)...)
	poscmd = append(poscmd, 0)
	poscmd = append(poscmd, 1)
	poscmd = append(poscmd, []byte{0, 0, 0, 0}...)
	poscmd = append(poscmd, []byte{0x00, 0x0C, 0x00, 0x03}...)
	poscmd = append(poscmd, []byte{0x01, 0x57, 0x8e, 0xA6}...)
	poscmd = append(poscmd, []byte{0x06, 0xca, 0x3c, 0x10}...)
	poscmd = append(poscmd, []byte{0x00, 0x30, 0x00, 0x00, 0x00, 0x00}...)
	poscmd = append(poscmd, addtime()...)
	poscmd = append(poscmd, []byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x00}...)
	poscmd = append(poscmd, []byte{0x03, 0x02, 0x00, 0x00}...)
	poscmd = append(poscmd, []byte{0x25, 0x04, 0x00, 0x00, 0x00, 0x00}...)
	poscmd = append(poscmd, []byte{0x30, 0x01, 0x13}...)
	poscmd = append(poscmd, []byte{0x31, 0x01, 0x12}...)
	poscmd = append(poscmd, []byte{0xe3, 0x10, 0x5a, 0x06, 0x02, 0x34, 0x02, 0x34, 0x01, 0x5f, 0x5b, 0x06, 0x01, 0x5f, 0x01, 0x5f, 0x01, 0x5f}...)
	poscmd = append(poscmd, []byte{0xe4, 0x06, 0x01, 0x19, 0x01, 0x1a, 0x01, 0x1b}...)
	checksumbyte := poscmd[1:len(poscmd)]
	checksum := CheckSum(checksumbyte, uint16(len(checksumbyte)))
	poscmd = append(poscmd, checksum)
	poscmd = append(poscmd, 0x7e)
	//poscmd := []byte{0x7e, 0x02, 0x00, 0x00, 0x40, 0x01, 0x38, 0x32, 0x35, 0x73, 0x52, 0x00, 0x0b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x03, 0x02, 0x42, 0x70, 0x2d, 0x06, 0xd8, 0x63, 0xa5, 0x00, 0x1f, 0x00, 0x00, 0x00, 0x00, 0x15, 0x09, 0x28, 0x18, 0x21, 0x26, 0x01, 0x04, 0x00, 0x00, 0x01, 0x25, 0x03, 0x02, 0x00, 0x00, 0x25, 0x04, 0x00, 0x00, 0x00, 0x00, 0x30, 0x01, 0x1a, 0x31, 0x01, 0x15, 0xe3, 0x08, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0xe4, 0x02, 0x00, 0x00, 0x6a, 0x7e}
	log.Printf("%x\n", poscmd)

	_, err := conn.Write(poscmd)
	if err != nil {
		log.Println(err.Error())
		go do(id, wg)

		return false
	}

	return true

}

var successauth uint32 = 0
var failauth uint32 = 0

var successpos uint32 = 0
var failpos uint32 = 0

func do(id string, wg *sync.WaitGroup) {
	wg.Add(1)
	srvaddr := "211.142.200.228:10054"
	//srvaddr := "192.168.2.111:9000"
	tcpaddr, _ := net.ResolveTCPAddr("tcp", srvaddr)

	conn, err := net.DialTCP("tcp", nil, tcpaddr)
	defer func() {
		wg.Done()
		conn.Close()
	}()
	if err != nil {
		log.Println(err.Error())
		return
	}

	sendauthcmd(id, conn)
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	//if buffer[17] == 0 {
	if err == nil {
		ticker := time.NewTicker(1 * 1e9)
		successauth++

		for {
			<-ticker.C
			if !sendposcmd(id, conn, wg) {
				return
			}
			conn.Read(buffer)
			if buffer[17] == 0 {
				successpos++
			} else {
				failpos++
			}
		}
	} else {
		log.Println("auth err")
		log.Println(err.Error())
		failauth++
	}

}

func main() {
	file, _ := os.OpenFile("./vehicles.txt", os.O_RDONLY, 0666)
	reader := bufio.NewReader(file)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	for {
		buf, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		time.Sleep(100000000)

		go do(string(buf), wg)

	}
	time.Sleep(1 * time.Second * 100)

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			<-ticker.C
			log.Printf("success auth %d, fail auth %d, pos success %d, pos fail %d", successauth, failauth, successpos, failpos)
		}
	}()
	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	wg.Wait()

}
