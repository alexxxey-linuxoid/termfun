package main

import (
	"log"

	"github.com/alexxxey-linuxoid/termfun/internal/menu"
	"github.com/alexxxey-linuxoid/termfun/internal/modules/night_sky"
	"github.com/alexxxey-linuxoid/termfun/internal/modules/space_shooter"
)

func main() {
	for {
		selected, err := menu.Run()
		if err != nil {
			log.Fatal(err)
		}

		switch selected {
		case "Night Sky":
			if err := night_sky.Run(); err != nil {
				log.Fatal(err)
			}

		case "Space Shooter":
			if err := space_shooter.Run(); err != nil {
				log.Fatal(err)
			}

		case "Exit", "":
			return
		}
	}
}
