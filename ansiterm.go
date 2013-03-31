// ansiterm.go
// see http://en.wikipedia.org/wiki/ANSI_escape_code
//		NOTE: won't work for older MSWin consoles - before DOS 2.0
//		NOTE: won't work for newer MSWin consoles - Win32 and later
//		Should work on your VT-100 if hardware still functions :-)
//		Works on (some/most?) xterms, including default GNOME consoles
//
//		Using "fmt" to send single bytes is probably not efficient
//		but unless you're sending massive data to screen you probably
//		won't notice...
//		It's primarily intended to make "status reports" easier to handle on
//		long running applications that don't warrant a full GUI workup.
//
//		See the demo program's Headline and StatusUpdate functions.
//		Because that's ALL it was intended to do I have implemented only
//		the ansi codes that were needed for that limited objective.
package ansiterm

// BUG(mdr): need to make so prompt length is ignored in "width"
// row,col is where the data field starts, prompt (if any) is adjusted to print at left of data

import (
	"fmt"
	"os"
	"sync"
)

const (
	ESC     = 033
	NORMAL  = 0
	INVERSE = 7
)

type ScreenForm struct {
	Fields map[string]*ScreenField
}

type ScreenField struct {
	tag         string // used as index in ScreenForm.Fields[tag]
	prompt      string
	msg         string
	row         int
	col         int
	width       int // constrains length of prompt (if any) + msg (if any)
	isInvisible bool
	inuse       sync.Mutex // intended to be useful in parallel operation
}

var (
	version = "ansiterm.go (c) 2013 David Rook released under Simplified BSD License"
)

// ----------------------------------------------------------------  ScreenForm

func (sf *ScreenForm) AddField(f *ScreenField) {
	if sf.Fields == nil {
		sf.Fields = make(map[string]*ScreenField)
	}
	sf.Fields[f.tag] = f
}

func (sf *ScreenForm) DeleteField(tag string) {
	if sf.Fields == nil {
		return
	}
	delete(sf.Fields, tag)
}

func (sf *ScreenForm) Draw() {
	ClearPage()
	for _, fld := range sf.Fields {
		fld.Draw()
	}
	MoveToRC(1, 1)
}

func (sf *ScreenForm) UpdateMsg(fieldName, msg string) {
	sf.Fields[fieldName].inuse.Lock()
	sf.Fields[fieldName].msg = msg
	sf.Fields[fieldName].inuse.Unlock()
}

// ---------------------------------------------------------------  ScreenField

func (f *ScreenField) SetTag(tag string) {
	f.inuse.Lock()
	f.tag = tag
	f.inuse.Unlock()
}

func (f *ScreenField) SetRCW(row, col, width int) {
	f.inuse.Lock()
	f.row, f.col, f.width = row, col, width
	f.inuse.Unlock()
}

func (f *ScreenField) SetPrompt(prompt string) {
	f.inuse.Lock()
	f.prompt = prompt
	f.inuse.Unlock()
}

func (f *ScreenField) Erase() {
	MoveToRC(f.row, f.col)
	Erase(f.width)
}

func (f *ScreenField) Draw() {
	if f.isInvisible {
		return
	}
	f.Erase()
	s := f.prompt + f.msg
	if len(s) > f.width {
		s = s[:f.width]
	}
	fmt.Printf("%s", s)
}

// --------------------------------------------------------------------

// erase whole page, leave cursor at 1,1
// 		ansi ED, special case n = 2
func ClearPage() {
	fmt.Printf("\033[2J")
	MoveToRC(1, 1)
}

// erase from cursor to end of line
// 		ansi EL specific case n = missing
func ClearLine() {
	fmt.Printf("\033[K")	
}

// ansi Query Position returns Esc[row;colR
// BUG(mdr): waits for Enter key since stdin is cooked
func QueryPosn() {
	fmt.Printf("\033[6n")
	var buf = make([]byte,20)
	nin,err := os.Stdin.Read(buf)	
	fmt.Printf("%s %d %v\n",buf,nin,err)
}

// ansi SCP
func SavePosn() {
	fmt.Printf("\033[s")
}

// ansi RCP
func RestorePosn() {
	fmt.Printf("\033[u")
}

func HideCursor() {
	fmt.Printf("\033[?25l")	
}

func ShowCursor() {
	fmt.Printf("\033[?25h")	
}

// BUG(mdr): not very efficient - Erase(n)
// erase N chars but dont move cursor position (clear field for printing)
func Erase(nchars int) {
	i := 0
	for nchars > 0 {
		nchars--
		fmt.Printf(" ")
		i++
	}
	for i > 0 {
		i--
		fmt.Printf("\b")
	}
}

// ansi HVP
func MoveToRC(row, col int) {
	fmt.Printf("\033[%d;%df", row, col)
}

// sugar for HVP
func MoveToXY(x, y int) {
	MoveToRC(y, x)
}

// 
func ResetTerm(attr int) {
	fmt.Printf("\033[1;80;0m") // restore normal attributes	
	if attr == NORMAL {
		return
	}
	fmt.Printf("033[1;1;%dm", attr)
}
