package loopback

import (
	"bytes"
	"os"

	"golang.zx2c4.com/wireguard/tun"
)

type Tun struct {
	events      chan tun.Event
	buf         *bytes.Buffer
	writeSignal chan struct{}
	readSignal  chan struct{}
	mtu         int
}

func CreateTun(mtu int) tun.Device {
	dev := &Tun{
		events:      make(chan tun.Event, 10),
		buf:         bytes.NewBuffer(nil),
		writeSignal: make(chan struct{}, 1),
		readSignal:  make(chan struct{}),
		mtu:         mtu,
	}
	dev.events <- tun.EventUp
	dev.writeSignal <- struct{}{}
	return dev
}

func (tun *Tun) File() *os.File {
	return nil
}

func (tun *Tun) Read(buffs [][]byte, sizes []int, offset int) (int, error) {
	_, ok := <-tun.readSignal
	if !ok {
		return 0, os.ErrClosed
	}
	n, err := tun.buf.Read(buffs[0][offset:])
	sizes[0] = n
	tun.writeSignal <- struct{}{}
	return 1, err
}

func (tun *Tun) Write(buffs [][]byte, offset int) (int, error) {
	if len(buffs) != 1 {
		panic("loopback: invalid batch size")
	}
	packet := buffs[0][offset:]
	if len(packet) == 0 {
		return 1, nil
	}
	_, ok := <-tun.writeSignal
	if !ok {
		return 0, os.ErrClosed
	}
	tun.buf.Reset()
	_, err := tun.buf.Write(packet)
	tun.readSignal <- struct{}{}
	return 1, err
}

func (tun *Tun) MTU() (int, error) {
	return tun.mtu, nil
}

func (tun *Tun) Name() (string, error) {
	return "loopback", nil
}

func (tun *Tun) Events() <-chan tun.Event {
	return tun.events
}

func (tun *Tun) Close() error {
	if tun.events != nil {
		close(tun.events)
	}

	// take out the write signal
	<-tun.writeSignal
	if tun.writeSignal != nil {
		close(tun.writeSignal)
	}
	if tun.readSignal != nil {
		close(tun.readSignal)
	}
	return nil
}

func (tun *Tun) BatchSize() int {
	return 1
}
