// Lines from wifly end \r\n
package gowifly

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus" // A replacement for the stdlib log
	"github.com/tarm/goserial"
	"io"
	"strings"
)

type WiFlyConnection struct {
	serialConn  *io.ReadWriteCloser
	commandMode bool
}

// TODO should return error
func (wifly *WiFlyConnection) EnterCommandMode() {
	log.Print("Entering command mode")
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

// Simply echos all data from the serial port to stdout
func echoSerial(s *io.ReadWriteCloser) {
	// http://stackoverflow.com/questions/17599232/reading-from-serial-port-with-while-loop
	reader := bufio.NewReader(*s)
	for true {
		reply, err := reader.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		m := strings.Replace(string(reply), "\n", "\\n", -1)
		m = strings.Replace(m, "\r", "\\r", -1)
		fmt.Println(string(m))
	}

}

// Simply echos all data from the serial port to stdout
// We can require no debug messages be printed at start
// Should have error checking in case we get off.
func inputHandler(s *io.ReadWriteCloser) {

}

// Returns a new WiFlyConnection connected to a serial port and normalized.
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
