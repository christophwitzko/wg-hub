package loopback

import (
	"os"

	"golang.zx2c4.com/wireguard/tun"
)

type Tun struct {
	events chan tun.Event
	queue  chan []byte
	mtu    int
}

func CreateTun(mtu int) tun.Device {
	dev := &Tun{
		events: make(chan tun.Event, 10),
		queue:  make(chan []byte),
		mtu:    mtu,
	}
	dev.events <- tun.EventUp
	return dev
}

func (tun *Tun) File() *os.File {
	return nil
}

func (tun *Tun) Read(buf []byte, offset int) (int, error) {
	data, ok := <-tun.queue
	if !ok {
		return 0, os.ErrClosed
	}
	copy(buf[offset:], data)
	return len(data), nil
}

func (tun *Tun) Write(buf []byte, offset int) (int, error) {
	packet := buf[offset:]
	if len(packet) == 0 {
		return 0, nil
	}
	tun.queue <- packet
	return len(buf), nil
}

func (tun *Tun) Flush() error {
	return nil
}

func (tun *Tun) MTU() (int, error) {
	return tun.mtu, nil
}

func (tun *Tun) Name() (string, error) {
	return "loopback", nil
}

func (tun *Tun) Events() chan tun.Event {
	return tun.events
}

func (tun *Tun) Close() error {
	if tun.events != nil {
		close(tun.events)
	}

	if tun.queue != nil {
		close(tun.queue)
	}
	return nil
}
