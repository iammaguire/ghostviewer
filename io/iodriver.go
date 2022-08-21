package io

func PassMessageToIODriver(m []byte) {
	//msg := string(m)
	//args := strings.Split(msg, ":")
	/*if strings.Contains(args[0], "MOUSE") {
		mX, _ := strconv.Atoi(args[1])
		mY, _ := strconv.Atoi(args[2])
		robotgo.Move(mX, mY)

		if args[0][0:2] == "LD" {
			robotgo.Toggle("left", "down")
		} else if args[0][0:2] == "RD" {
			robotgo.Toggle("right", "down")
		} else if args[0][0:2] == "LU" {
			robotgo.Toggle("left", "up")
		} else if args[0][0:2] == "RU" {
			robotgo.Toggle("right", "up")
		}
	} else if args[0] == "KEY" {
		if _, err := strconv.Atoi(string(args[2])); err == nil { // checks if idx is valid, failsafe measure
			fmt.Println(msg)

			mK, _ := strconv.Atoi(string(args[1]))
			mKind := uint8(mK)
			kChar := string(args[3])
			switch mKind {
			case gohook.KeyDown:
				robotgo.KeyDown(string(kChar))
				robotgo.KeyUp(string(kChar))
			} This is another test. hope it works okl. This is pushingi t I think. it's starting to get a little laggy.
		}
	}*/
}
