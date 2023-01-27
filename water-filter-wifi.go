package main

import (
	"log"
	"os/exec"
	"time"

	rpio "github.com/warthog618/gpiod"
)

const (
	remoteNo   int = 26
	solenoidNo int = 4
	filepath       = "runtime"
	soundPath      = "/home/pi/water-filter/chirp.wav"
	msg            = "click"
)

var runtime uint8 = 40

/*
remote :
	5v pin no 4 brown
	gnd pin 6 black
	signal pin no color bcm-26

solenoid relay :
    5v pin no 2 blue
    gnd pin no 9 green
	solenoid pin no color bcm-4

	TODO : find pins for chirper

	U file 8bit 8khz
*/

type wfilter struct {
	remote   *rpio.Line
	solenoid *rpio.Line
	clicks   *chan string
	command  *exec.Cmd
}

func newWFilter() *wfilter {
	channel := make(chan string)

	return &wfilter{clicks: &channel}
}

func (w *wfilter) setup() error {
	cmd := exec.Command("aplay", soundPath)

	remotePin, err := rpio.RequestLine("gpiochip0", remoteNo,
		rpio.WithPullDown,
		rpio.WithRisingEdge,
		rpio.WithEventHandler(w.handleClick))
	if err != nil {
		return err
	}

	solenoidPin, err := rpio.RequestLine("gpiochip0", solenoidNo, rpio.AsOutput(0))
	if err != nil {
		return err
	}

	w.remote = remotePin
	w.solenoid = solenoidPin
	w.command = cmd

	return nil
}

func (w *wfilter) chirp() error {
	err := w.command.Run()
	return err
}

func (w *wfilter) handleClick(_ rpio.LineEvent) {
	select {
	case *w.clicks <- msg:
		// sucessfully sent
	default:
		// channel full
	}
}

func (w *wfilter) runWater() {
	w.solenoid.SetValue(1)

	err := w.chirp()
	if err != nil {
		log.Println("error while chirping", err)
	}

	time.Sleep(time.Second * time.Duration(runtime))
	// TODO: account for time taken by chirp
	w.solenoid.SetValue(0)
}

func (w *wfilter) clearClicks() {
	select {
	case _ = <-*w.clicks:
		// cleared
	default:
		// nothing to clear
	}
}

func (w *wfilter) waitForClick() {
	_ = <-*w.clicks
}

func main() {
	wf := newWFilter()
	err := wf.setup()
	if err != nil {
		log.Println("error while setting up", err)
		return
	}

	for {
		wf.waitForClick()
		wf.runWater()
		wf.clearClicks()
	}
}
