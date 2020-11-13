/*
 * @Author: gitsrc
 * @Date: 2020-11-10 19:24:27
 * @LastEditors: gitsrc
 * @LastEditTime: 2020-11-13 15:32:23
 * @FilePath: /influx-proxy-gitsrc/service/udp.go
 */

package service

import (
	"log"
	"net"

	"github.com/chengshiwen/influx-proxy/backend"
	"github.com/chengshiwen/influx-proxy/transfer"
	"github.com/oxtoacart/bpool"
	"github.com/panjf2000/ants/v2"
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
	var bufpool *bpool.BytePool
	// go func() {
	// 	for {
	// 		time.Sleep(time.Millisecond * 100)
	// 		log.Println("count", bufpool.NumPooled())

	// 	}
	// }()

	//create go pool
	p, err := ants.NewPool(10000)
	if err != nil {
		log.Println(err)
		return
	}
	defer p.Release() //release go pool

	//create
	bufpool = bpool.NewBytePool(125, 1024)

	for {
		buf := bufpool.Get() //内存池缓冲区获取

		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			bufpool.Put(buf)
			log.Println(err)
			continue
		}
		ants.Submit(func() {
			us.process(pc, addr, bufpool, buf, n) //内部包含内存释放
		})
	}
}

func (us *UDPService) process(pc net.PacketConn, addr net.Addr, bufpool *bpool.BytePool, buf []byte, n int) {
	defer bufpool.Put(buf) //释放内存进入池中
	if us.WriteTracing {
		log.Printf("write: [%s %s]\n", us.UDPDatabase, buf[:n])
	}
	err := us.ip.Write(buf[:n], us.UDPDatabase, "ns")
	if err != nil {
		log.Println(err)
	}
	// atomic.AddUint64(&count, 1)
	// log.Println(atomic.LoadUint64(&count))
}
