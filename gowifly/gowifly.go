// Lines from wifly end \r\n
// TODO write tests for scanner functions as well
package gowifly

import (
	"io"
	"strings"

	log "github.com/Sirupsen/logrus" // A replacement for the stdlib log
	"github.com/lelandmiller/SCcomm/packet"
	"github.com/tarm/goserial"
)

type WiFlyConnection struct {
	serialConn  *io.ReadWriteCloser
	commandMode bool
}

// TODO should return error
func (wifly *WiFlyConnection) EnterCommandMode() {
	//log.Print("Entering command mode")
	wifly.write("$$$")

}

// TODO should return error
// Does not expect the command to contain newlines or carriage returns
func (wifly *WiFlyConnection) write(command string) {
	_, err := (*wifly.serialConn).Write([]byte(command))
	if err != nil {
		log.Panic(err)
	}
}
func (wifly *WiFlyConnection) Write(command string) {
	wifly.write(command)
}
func (wifly *WiFlyConnection) Stream() *io.ReadWriteCloser {
	return wifly.serialConn
}

// TODO should return error
// Does not expect the command to contain newlines or carriage returns
func (wifly *WiFlyConnection) WriteCommand(command string) {
	command = strings.Join([]string{command, "\r"}, "")
	wifly.write(command)
}
func (wifly *WiFlyConnection) WriteRawPacket(p *packet.Packet) {
	wifly.write(string(p.WireFormat()))
}

// Returns a new WiFlyConnection connected to a serial port and normalized.
// TODO normalize
func NewWiFlyConnection() *WiFlyConnection {
	conn := new(WiFlyConnection)
	conn.SerialConnect()
	return conn
}

// Connects the WiFlyConnection to a serial port.
func (w *WiFlyConnection) SerialConnect() {
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Panic(err)
	}
	//w.buffReader = bufio.NewReader(s)
	w.serialConn = &s
}
