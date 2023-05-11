package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gen2brain/go-mpv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Provide a filename on the command line.")
		return
	}

	m := mpv.Create()
	err := m.SetOptionString("input-default-bindings", "yes")
	if err != nil {
		log.Fatal(err)
	}

	m.SetOptionString("input-vo-keyboard", "yes")
	m.SetOption("osc", mpv.FORMAT_FLAG, true)

	err = m.Initialize()
	if err != nil {
		log.Fatal(err)
	}

	err = m.Command([]string{"loadfile", os.Args[1]})
	if err != nil {
		log.Fatal(err)
	}

	for {
		e := m.WaitEvent(10000)
		log.Println("event:", e.Event_Id)
		if e.Event_Id == mpv.EVENT_SHUTDOWN || e.Event_Id == mpv.EVENT_IDLE {
			break
		}
	}

	m.TerminateDestroy()
}
