package service

import (
	"log"
	"net"

	"github.com/chengshiwen/influx-proxy/backend"
	"github.com/chengshiwen/influx-proxy/transfer"
)

//UDPService is UDP server
type UDPService struct {
	ip           *backend.Proxy
	tx           *transfer.Transfer
	WriteTracing bool
	UDPBind      string //UDP监控地址
	UDPDatabase  string //UDP数据库
}

// NewUDPService is create udp server object
func NewUDPService(cfg *backend.ProxyConfig) (us *UDPService) { // nolint:golint
	ip := backend.NewProxy(cfg) //create influx proxy object by config
	us = &UDPService{
		ip:           ip,
		tx:           transfer.NewTransfer(cfg, ip.Circles),
		UDPBind:      cfg.UDPBind,
		UDPDatabase:  cfg.UDPDataBase,
		WriteTracing: cfg.WriteTracing,
	}
	return
}

//ListenAndServe is UDP server Bind Handle
func (us *UDPService) ListenAndServe() (err error) {
	pc, err := net.ListenPacket("udp", us.UDPBind)
	if err != nil {
		return
	}
	defer pc.Close()

	log.Printf("UDP service start on DB [%s], listen %s", us.UDPDatabase, us.UDPBind)
	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		go us.process(pc, addr, buf[:n])
	}
}

func (us *UDPService) process(pc net.PacketConn, addr net.Addr, buf []byte) {

	if us.WriteTracing {
		log.Printf("write: [%s %s]\n", us.UDPDatabase, buf)
	}
	err := us.ip.Write(buf, us.UDPDatabase, "ns")
	if err != nil {
		log.Println(err)
	}
	// atomic.AddUint64(&count, 1)
	// log.Println(atomic.LoadUint64(&count))
}
