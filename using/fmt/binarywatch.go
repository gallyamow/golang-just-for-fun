package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type binTime struct {
	h1, h2 byte
	m1, m2 byte
	s1, s2 byte
}

func currBinTime() *binTime {
	now := time.Now()
	return &binTime{
		h1: byte(now.Hour() / 10),
		h2: byte(now.Hour() % 10),
		m1: byte(now.Minute() / 10),
		m2: byte(now.Minute() % 10),
		s1: byte(now.Second() / 10),
		s2: byte(now.Second() % 10),
	}
}

func (b *binTime) String() string {
	return fmt.Sprintf("%d%d:%d%d:%d%d", b.h1, b.h2, b.m1, b.m2, b.s1, b.s2)
}

// pointDisplay12 represents time in 24-hours format
type pointDisplay12 struct {
	h1 [2]bool // 0-2
	h2 [4]bool // 0-8
	m1 [3]bool // 0-4
	m2 [4]bool // 0-8
	s1 [3]bool // 0-3
	s2 [4]bool // 0-8
}

func (d *pointDisplay12) applyTime(b *binTime) {
	// способ через &2^
	d.h1[0] = b.h1&1 == 1
	d.h1[1] = b.h1&2 == 2
	d.h2[0] = b.h2&1 == 1
	d.h2[1] = b.h2&2 == 2
	d.h2[2] = b.h2&4 == 4
	d.h2[3] = b.h2&8 == 8

	// способ через сдвиг и &1
	d.m1[0] = b.m1>>0&1 == 1
	d.m1[1] = b.m1>>1&1 == 2
	d.m1[2] = b.m1>>2&1 == 4

	d.m2[0] = b.m2&1 == 1
	d.m2[1] = b.m2&2 == 2
	d.m2[2] = b.m2&4 == 4
	d.m2[3] = b.m2&8 == 8

	d.s1[0] = b.s1&1 == 1
	d.s1[1] = b.s1&2 == 2
	d.s1[2] = b.s1&4 == 4

	d.s2[0] = b.s2&1 == 1
	d.s2[1] = b.s2&2 == 2
	d.s2[2] = b.s2&4 == 4
	d.s2[3] = b.s2&8 == 8
}

const EmptyPoint = ' '
const FullPoint = '●'
const NonePoint = '○'

func resolvePointRune(zi int, column []bool) rune {
	p := EmptyPoint
	//fmt.Println(len(column), column, zi)
	if len(column) > zi {
		if column[zi] {
			p = FullPoint
		} else {
			p = NonePoint
		}
	}
	return p
}

func (d *pointDisplay12) Print(cb *binTime) {
	d.applyTime(cb)

	for i := 0; i < 4; i++ {
		zi := 4 - i - 1

		var p rune
		for j := 0; j < 6; j++ {
			switch j {
			// hours
			case 0:
				p = resolvePointRune(zi, d.h1[:])
			case 1:
				p = resolvePointRune(zi, d.h2[:])
			// minutes
			case 2:
				p = resolvePointRune(zi, d.m1[:])
			case 3:
				p = resolvePointRune(zi, d.m2[:])
			// seconds
			case 4:
				p = resolvePointRune(zi, d.s1[:])
			case 5:
				p = resolvePointRune(zi, d.s2[:])
			}

			fmt.Printf("%3c", p)
		}
		fmt.Println()
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // os.Interrupt = syscall.SIGINT

	display := &pointDisplay12{}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c:
			return
		default:
			cb := currBinTime()

			// clear screen & move cursor to top left
			fmt.Print("\033[H\033[2J")

			fmt.Printf("%s%s\n", strings.Repeat(" ", 7), cb)
			display.applyTime(cb)
			display.Print(cb)

			<-ticker.C
		}
	}
}
