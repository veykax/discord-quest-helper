//go:build linux

package stub

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func pad(n int) int {
	return (4 - (n % 4)) % 4
}

func readXAuthority() (name, data []byte, err error) {
	candidates := []string{
		os.Getenv("XAUTHORITY"),
		"/run/flatpak/Xauthority",
	}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".Xauthority"))
	}
	// Also check /run/user/<uid>/xauth_*
	if matches, err := filepath.Glob(fmt.Sprintf("/run/user/%d/xauth_*", os.Getuid())); err == nil {
		candidates = append(candidates, matches...)
	}

	var authPath string
	for _, p := range candidates {
		if p != "" {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() && fi.Size() > 0 {
				authPath = p
				break
			}
		}
	}
	if authPath == "" {
		return nil, nil, fmt.Errorf("no xauthority file found")
	}

	f, err := os.Open(authPath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	for {
		var fam uint16
		if err := binary.Read(f, binary.BigEndian, &fam); err != nil {
			break
		}
		var alen uint16
		if err := binary.Read(f, binary.BigEndian, &alen); err != nil {
			break
		}
		addr := make([]byte, alen)
		if _, err := io.ReadFull(f, addr); err != nil {
			break
		}

		var nlen uint16
		if err := binary.Read(f, binary.BigEndian, &nlen); err != nil {
			break
		}
		num := make([]byte, nlen)
		if _, err := io.ReadFull(f, num); err != nil {
			break
		}

		var nameLen uint16
		if err := binary.Read(f, binary.BigEndian, &nameLen); err != nil {
			break
		}
		protoName := make([]byte, nameLen)
		if _, err := io.ReadFull(f, protoName); err != nil {
			break
		}

		var dataLen uint16
		if err := binary.Read(f, binary.BigEndian, &dataLen); err != nil {
			break
		}
		protoData := make([]byte, dataLen)
		if _, err := io.ReadFull(f, protoData); err != nil {
			break
		}

		if string(protoName) == "MIT-MAGIC-COOKIE-1" {
			return protoName, protoData, nil
		}
	}

	return nil, nil, fmt.Errorf("no MIT-MAGIC-COOKIE-1 found in %s", authPath)
}

func Run(title string) {
	if title == "" {
		title = "Game"
	}
	exeName := filepath.Base(os.Args[0])

	display := os.Getenv("DISPLAY")
	if display == "" {
		display = ":0"
	}
	dispNum := "0"
	if len(display) > 1 && display[0] == ':' {
		dispNum = display[1:]
	}
	sockPath := fmt.Sprintf("/tmp/.X11-unix/X%s", dispNum)
	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		fmt.Printf("[Stub] Dial X11 err: %v\n", err)
		fallbackWait()
		return
	}
	defer conn.Close()

	protoName, protoData, err := readXAuthority()
	if err != nil {
		fmt.Printf("[Stub] Xauthority err: %v\n", err)
		fallbackWait()
		return
	}

	// 1. Send Connection Setup Request
	buf := new(bytes.Buffer)
	buf.WriteByte('l') // Little-endian
	buf.WriteByte(0)
	binary.Write(buf, binary.LittleEndian, uint16(11)) // Major 11
	binary.Write(buf, binary.LittleEndian, uint16(0))  // Minor 0
	binary.Write(buf, binary.LittleEndian, uint16(len(protoName)))
	binary.Write(buf, binary.LittleEndian, uint16(len(protoData)))
	binary.Write(buf, binary.LittleEndian, uint16(0)) // pad
	buf.Write(protoName)
	buf.Write(make([]byte, pad(len(protoName))))
	buf.Write(protoData)
	buf.Write(make([]byte, pad(len(protoData))))

	if _, err := conn.Write(buf.Bytes()); err != nil {
		fallbackWait()
		return
	}

	// 2. Read Setup Reply Header
	hdr := make([]byte, 8)
	if _, err := io.ReadFull(conn, hdr); err != nil || hdr[0] != 1 {
		fallbackWait()
		return
	}

	addLen := binary.LittleEndian.Uint16(hdr[6:8])
	extra := make([]byte, int(addLen)*4)
	if _, err := io.ReadFull(conn, extra); err != nil {
		fallbackWait()
		return
	}

	resBase := binary.LittleEndian.Uint32(extra[4:8])
	vLen := binary.LittleEndian.Uint16(extra[16:18])
	nFormats := int(extra[21])
	screenOffset := 32 + int(vLen) + pad(int(vLen)) + nFormats*8
	if len(extra) < screenOffset+4 {
		fallbackWait()
		return
	}
	rootWindow := binary.LittleEndian.Uint32(extra[screenOffset : screenOffset+4])

	winID := resBase | 1
	gcID := resBase | 2

	// InternAtom helper
	internAtom := func(name string) uint32 {
		b := new(bytes.Buffer)
		b.WriteByte(16) // InternAtom opcode
		b.WriteByte(0)  // only_if_exists = false
		nameLen := len(name)
		p := pad(nameLen)
		binary.Write(b, binary.LittleEndian, uint16(2+(nameLen+p)/4))
		binary.Write(b, binary.LittleEndian, uint16(nameLen))
		binary.Write(b, binary.LittleEndian, uint16(0))
		b.WriteString(name)
		b.Write(make([]byte, p))
		conn.Write(b.Bytes())

		r := make([]byte, 32)
		io.ReadFull(conn, r)
		return binary.LittleEndian.Uint32(r[8:12])
	}

	atomString := uint32(31)   // XA_STRING
	atomCardinal := uint32(6) // XA_CARDINAL
	atomAtom := uint32(4)     // XA_ATOM
	atomWmName := uint32(39)   // XA_WM_NAME
	atomWmClass := uint32(67)  // XA_WM_CLASS
	atomWmHints := uint32(35)  // XA_WM_HINTS
	atomNetWmName := internAtom("_NET_WM_NAME")
	atomUtf8String := internAtom("UTF8_STRING")
	atomNetWmPid := internAtom("_NET_WM_PID")
	atomNetWmWinType := internAtom("_NET_WM_WINDOW_TYPE")
	atomNetWmWinTypeNormal := internAtom("_NET_WM_WINDOW_TYPE_NORMAL")
	atomWmProtocols := internAtom("WM_PROTOCOLS")
	atomWmDeleteWindow := internAtom("WM_DELETE_WINDOW")
	atomWmChangeState := internAtom("WM_CHANGE_STATE")

	// 3. CreateWindow (Opcode 1)
	// value mask: CWBackPixel (0x2) | CWEventMask (0x800) = 0x802 (2 values)
	// EventMask: ExposureMask (0x8000) | StructureNotifyMask (0x20000) = 0x28000
	cw := new(bytes.Buffer)
	cw.WriteByte(1) // Opcode 1
	cw.WriteByte(0) // depth 0 = copy from parent
	binary.Write(cw, binary.LittleEndian, uint16(10)) // 8 + 2
	binary.Write(cw, binary.LittleEndian, winID)
	binary.Write(cw, binary.LittleEndian, rootWindow)
	binary.Write(cw, binary.LittleEndian, int16(-32000)) // x (off-screen)
	binary.Write(cw, binary.LittleEndian, int16(-32000)) // y (off-screen)
	binary.Write(cw, binary.LittleEndian, uint16(360)) // width
	binary.Write(cw, binary.LittleEndian, uint16(180)) // height
	binary.Write(cw, binary.LittleEndian, uint16(0))   // border width
	binary.Write(cw, binary.LittleEndian, uint16(0))   // class CopyFromParent
	binary.Write(cw, binary.LittleEndian, uint32(0))   // visual CopyFromParent
	binary.Write(cw, binary.LittleEndian, uint32(0x802)) // value mask
	binary.Write(cw, binary.LittleEndian, uint32(0x1e1f22)) // CWBackPixel (Discord dark gray)
	binary.Write(cw, binary.LittleEndian, uint32(0x28000))  // CWEventMask
	conn.Write(cw.Bytes())

	// CreateGC (Opcode 55)
	// value mask: GCForeground (0x4) | GCBackground (0x8) = 0xc
	cgc := new(bytes.Buffer)
	cgc.WriteByte(55) // Opcode 55
	cgc.WriteByte(0)
	binary.Write(cgc, binary.LittleEndian, uint16(6)) // 4 + 2
	binary.Write(cgc, binary.LittleEndian, gcID)
	binary.Write(cgc, binary.LittleEndian, winID)
	binary.Write(cgc, binary.LittleEndian, uint32(0xc))
	binary.Write(cgc, binary.LittleEndian, uint32(0xffffff)) // foreground white
	binary.Write(cgc, binary.LittleEndian, uint32(0x1e1f22)) // background dark gray
	conn.Write(cgc.Bytes())

	changeProp := func(property, propType uint32, format byte, data []byte) {
		b := new(bytes.Buffer)
		b.WriteByte(18) // ChangeProperty opcode
		b.WriteByte(0)  // Replace
		p := pad(len(data))
		binary.Write(b, binary.LittleEndian, uint16(6+(len(data)+p)/4))
		binary.Write(b, binary.LittleEndian, winID)
		binary.Write(b, binary.LittleEndian, property)
		binary.Write(b, binary.LittleEndian, propType)
		b.WriteByte(format)
		b.WriteByte(0)
		b.WriteByte(0)
		b.WriteByte(0)
		numUnits := uint32(len(data))
		if format == 32 {
			numUnits = uint32(len(data) / 4)
		} else if format == 16 {
			numUnits = uint32(len(data) / 2)
		}
		binary.Write(b, binary.LittleEndian, numUnits)
		b.Write(data)
		b.Write(make([]byte, p))
		conn.Write(b.Bytes())
	}

	// 4. Set Properties for EWMH / Window Manager / Discord Detection
	changeProp(atomWmName, atomString, 8, []byte(title))
	changeProp(atomNetWmName, atomUtf8String, 8, []byte(title))

	// WM_CLASS: "<instance>\0<class>\0"
	classBytes := []byte(fmt.Sprintf("%s\x00%s\x00", exeName, exeName))
	changeProp(atomWmClass, atomString, 8, classBytes)

	// _NET_WM_PID: current process PID
	pidBuf := new(bytes.Buffer)
	binary.Write(pidBuf, binary.LittleEndian, uint32(os.Getpid()))
	changeProp(atomNetWmPid, atomCardinal, 32, pidBuf.Bytes())

	// _NET_WM_WINDOW_TYPE: _NET_WM_WINDOW_TYPE_NORMAL
	typeBuf := new(bytes.Buffer)
	binary.Write(typeBuf, binary.LittleEndian, atomNetWmWinTypeNormal)
	changeProp(atomNetWmWinType, atomAtom, 32, typeBuf.Bytes())

	// WM_PROTOCOLS: WM_DELETE_WINDOW
	protoBuf := new(bytes.Buffer)
	binary.Write(protoBuf, binary.LittleEndian, atomWmDeleteWindow)
	changeProp(atomWmProtocols, atomAtom, 32, protoBuf.Bytes())

	// WM_HINTS: IconicState
	hintsBuf := new(bytes.Buffer)
	binary.Write(hintsBuf, binary.LittleEndian, uint32(1|2)) // StateHint | IconPositionHint
	binary.Write(hintsBuf, binary.LittleEndian, uint32(0))   // input
	binary.Write(hintsBuf, binary.LittleEndian, uint32(3))   // initial_state = IconicState
	binary.Write(hintsBuf, binary.LittleEndian, uint32(0))   // icon_pixmap
	binary.Write(hintsBuf, binary.LittleEndian, uint32(0))   // icon_window
	binary.Write(hintsBuf, binary.LittleEndian, int32(0))    // icon_x
	binary.Write(hintsBuf, binary.LittleEndian, int32(0))    // icon_y
	binary.Write(hintsBuf, binary.LittleEndian, uint32(0))   // icon_mask
	binary.Write(hintsBuf, binary.LittleEndian, uint32(0))   // window_group
	changeProp(atomWmHints, atomWmHints, 32, hintsBuf.Bytes())

	// 5. MapWindow (Opcode 8)
	conn.Write([]byte{8, 0, 2, 0, byte(winID), byte(winID >> 8), byte(winID >> 16), byte(winID >> 24)})

	// 6. Send ClientMessage WM_CHANGE_STATE to root to minimize to taskbar
	cm := new(bytes.Buffer)
	cm.WriteByte(25) // SendEvent
	cm.WriteByte(0)  // propagate = false
	binary.Write(cm, binary.LittleEndian, uint16(11))
	binary.Write(cm, binary.LittleEndian, rootWindow)
	binary.Write(cm, binary.LittleEndian, uint32(0x100000|0x200000)) // SubstructureRedirect | SubstructureNotify
	cm.WriteByte(33)                                                 // ClientMessage
	cm.WriteByte(32)                                                 // format 32
	binary.Write(cm, binary.LittleEndian, uint16(0))
	binary.Write(cm, binary.LittleEndian, winID)
	binary.Write(cm, binary.LittleEndian, atomWmChangeState)
	binary.Write(cm, binary.LittleEndian, uint32(3)) // IconicState
	binary.Write(cm, binary.LittleEndian, uint32(0))
	binary.Write(cm, binary.LittleEndian, uint32(0))
	binary.Write(cm, binary.LittleEndian, uint32(0))
	binary.Write(cm, binary.LittleEndian, uint32(0))
	conn.Write(cm.Bytes())

	// Round-trip sync
	conn.Write([]byte{43, 0, 1, 0}) // GetInputFocus
	syncRep := make([]byte, 32)
	io.ReadFull(conn, syncRep)

	drawText := func(x, y int16, text string) {
		if len(text) > 255 {
			text = text[:255]
		}
		tBuf := new(bytes.Buffer)
		tBuf.WriteByte(76) // ImageText8
		tBuf.WriteByte(byte(len(text)))
		p := pad(len(text))
		binary.Write(tBuf, binary.LittleEndian, uint16(4+(len(text)+p)/4))
		binary.Write(tBuf, binary.LittleEndian, winID)
		binary.Write(tBuf, binary.LittleEndian, gcID)
		binary.Write(tBuf, binary.LittleEndian, x)
		binary.Write(tBuf, binary.LittleEndian, y)
		tBuf.WriteString(text)
		tBuf.Write(make([]byte, p))
		conn.Write(tBuf.Bytes())
	}

	drawAll := func() {
		var asciiTitle []byte
		for _, r := range title {
			if r < 128 {
				asciiTitle = append(asciiTitle, byte(r))
			} else {
				asciiTitle = append(asciiTitle, '?')
			}
		}
		drawText(30, 50, string(asciiTitle))
		drawText(30, 80, "(Discord Quest Simulation)")
		drawText(30, 115, "Leave minimized or close to stop.")
	}

	drawAll()

	exitChan := make(chan struct{})

	// Background event listener (e.g. WM_DELETE_WINDOW when user closes window, Expose when restored)
	go func() {
		ev := make([]byte, 32)
		for {
			if _, err := io.ReadFull(conn, ev); err != nil {
				close(exitChan)
				return
			}
			evType := ev[0] & 0x7f
			switch evType {
			case 12: // Expose: redraw window contents when unminimized or restored
				drawAll()
			case 33: // ClientMessage
				msgAtom := binary.LittleEndian.Uint32(ev[8:12])
				data0 := binary.LittleEndian.Uint32(ev[12:16])
				if msgAtom == atomWmProtocols && data0 == atomWmDeleteWindow {
					close(exitChan)
					return
				}
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
	case <-exitChan:
	}

	// DestroyWindow (Opcode 4)
	conn.Write([]byte{4, 0, 2, 0, byte(winID), byte(winID >> 8), byte(winID >> 16), byte(winID >> 24)})
	time.Sleep(50 * time.Millisecond)
}

func fallbackWait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
