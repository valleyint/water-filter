package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/aerth/playwav"
	rpio "github.com/warthog618/gpiod"
)

const (
	remoteNo   int = 26
	solenoidNo int = 4
	bellNo     int = 2
	// TODO:get actuall pin numbers from hardware
	filepath = "runtime"
)

var runtime uint8 = 40

/*
comon always high
nO normally open
// TODO:get new pin numbers from bash code
remote :
	5v pin no 4 brown
	gnd pin 6 black
	normally open pin 8 white bcm-14
	common pin no 10 grey bcm-15

relay :
    5v pin no 2 blue
    gnd pin no 9 green
	solenoid pin no  white 7  bcm-4
*/

type wfilter struct {
	listner    *net.TCPListener
	remote     *rpio.Line
	solenoid   *rpio.Line
	file       *os.File
	bell       *rpio.Line
	readWriter *bufio.ReadWriter
}

func (w *wfilter) intiFile() error {
	ourFile, err := os.Open(filepath)
	if err != nil {
		return err
	}

	ourReader := bufio.NewReader(ourFile)
	ourWriter := bufio.NewWriter(ourFile)
	ourReadWriter := bufio.NewReadWriter(ourReader, ourWriter)

	runtime, err = ourReader.ReadByte()
	if err != nil {
		return err
	}

	w.file = ourFile
	w.readWriter = ourReadWriter
	return nil
}

func setup() (*wfilter, error) {
	// addr, err := net.ResolveTCPAddr("tcp", myAddr)
	// if err != nil {
	// 	return nil, err
	// }

	w := wfilter{}
	/*
		l, err := net.ListenTCP("tcp", nil)
		if err != nil {
			return nil, err
		}
	*/
	// remote, err := rpio.NewChip("remote")
	// if err != nil {
	// 	return nil, err
	// }

	// solenoidChip, err := rpio.NewChip("solenoid")
	// if err != nil {
	// 	return nil, err
	// }

	// bellChip, err := rpio.NewChip("bell")
	// if err != nil {
	// 	return nil, err
	// }

	remotePin, err := rpio.RequestLine("gpiochip0", remoteNo,
		rpio.WithPullDown,
		rpio.WithRisingEdge,
		rpio.WithEventHandler(w.toRunWater))
	if err != nil {
		return nil, err
	}

	// nO, err := rpio.RequestLine("remote", remoteNoNO, gpiod.WithPullDown, gpiod.AsInput)
	// if  err != nil {
	// 	return nil, err
	// }

	solenoidPin, err := rpio.RequestLine("gpiochip0", solenoidNo, rpio.AsOutput(0))
	if err != nil {
		return nil, err
	}

	bellPin, err := rpio.RequestLine("gpiochip0", bellNo, rpio.AsOutput(1))
	if err != nil {
		return nil, err
	}

	w.remote = remotePin
	w.solenoid = solenoidPin
	w.bell = bellPin
	//	w.listner = l

	return &w, nil
}

func chirp() error {
	playwav.FromFile("/home/pi/waterfilter/chirp.wav")
	return nil
}

func (w *wfilter) handleRequest() error {
	for {
		conn, err := w.listner.Accept()
		if err != nil {
			return err
		}

		newValArr := []byte{0}
		conn.Read(newValArr)
		runtime = newValArr[0]

		n, err := w.file.WriteAt(newValArr, 0)
		if err != nil || n != 0 {
			conn.Close()
			return errors.New("didnt write")
		}
		conn.Close()
	}
}

func (w *wfilter) toRunWater(line rpio.LineEvent) {
	fmt.Println("got run command")
	w.runWater()
}

func (w *wfilter) runWater() {
	go chirp()
	w.solenoid.SetValue(1)
	time.Sleep(time.Second * time.Duration(runtime))
	w.solenoid.SetValue(0)
}

// func runner() {
// 	f, err := setup()
// 	if err != nil {
// 		panic(err)
// 	}

// 	go f.handleRequest()

// 	for {
// 		if f.toRunWater() {
// 			f.runWater()
// 			time.Sleep(time.Second * 10)
// 		}
// 		time.Sleep(time.Second)
// 	}
// }

func main() {
	_, err := setup()
	if err != nil {
		panic(err)
	}
	fmt.Println("finished setup")

	//	err = f.handleRequest()
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(10000)
	}
}
