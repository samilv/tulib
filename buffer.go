package tulib

import "github.com/nsf/termbox-go"
import "unicode/utf8"

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

type Buffer struct {
	Cells []termbox.Cell
	Width int
	Height int
}

func NewBuffer(w, h int) Buffer {
	return Buffer{
		Cells: make([]termbox.Cell, w*h),
		Width: w,
		Height: h,
	}
}

func TermboxBuffer() Buffer {
	w, h := termbox.Size()
	return Buffer{
		Cells: termbox.CellBuffer(),
		Width: w,
		Height: h,
	}
}

// Fills an area which is an intersection between buffer and 'dest' with 'proto'.
func (this *Buffer) Fill(dest Rect, proto termbox.Cell) {
	this.unsafe_fill(dest.Intersection(this.Rect()), proto)
}

// Resizes the Buffer, buffer contents are invalid after the resize.
func (this *Buffer) Resize(nw, nh int) {
	this.Width = nw
	this.Height = nh

	nsize := nw * nh
	if nsize <= cap(this.Cells) {
		this.Cells = this.Cells[:nsize]
	} else {
		this.Cells = make([]termbox.Cell, nsize)
	}
}

// Unsafe part of the fill operation, doesn't check for bounds.
func (this *Buffer) unsafe_fill(dest Rect, proto termbox.Cell) {
	stride := this.Width - dest.Width
	off := dest.Y * this.Width + dest.X
	for y := 0; y < dest.Height; y++ {
		for x := 0; x < dest.Width; x++ {
			this.Cells[off] = proto
			off++
		}
		off += stride
	}
}

func (this *Buffer) Rect() Rect {
	return Rect{0, 0, this.Width, this.Height}
}

// draws from left to right, 'off' is the beginning position
// (DrawLabel uses that method)
func (this *Buffer) draw_n_first_runes(off, n int, params *LabelParams, text string) {
	for _, r := range text {
		if n <= 0 {
			break
		}
		this.Cells[off] = termbox.Cell{
			Ch: r,
			Fg: params.Fg,
			Bg: params.Bg,
		}
		off++
		n--
	}
}

// draws from right to left, 'off' is the end position
// (DrawLabel uses that method)
func (this *Buffer) draw_n_last_runes(off, n int, params *LabelParams, text string) {
	for n > 0 {
		r, size := utf8.DecodeLastRuneInString(text)
		this.Cells[off] = termbox.Cell{
			Ch: r,
			Fg: params.Fg,
			Bg: params.Bg,
		}
		text = text[:len(text)-size]
		off--
		n--
	}
}

type LabelParams struct {
	Fg termbox.Attribute
	Bg termbox.Attribute
	Align Alignment
	Ellipsis rune
	CenterEllipsis bool
}

var DefaultLabelParams = LabelParams{
	termbox.ColorDefault,
	termbox.ColorDefault,
	AlignLeft,
	'…',
	false,
}

func skip_n_runes(x string, n int) string {
	if n <= 0 {
		return x
	}

	for n > 0 {
		_, size := utf8.DecodeRuneInString(x)
		x = x[size:]
		n--
	}
	return x
}

func (this *Buffer) DrawLabel(dest Rect, params *LabelParams, text string) {
	if dest.Height != 1 {
		dest.Height = 1
	}

	dest = dest.Intersection(this.Rect())
	if dest.Height == 0 || dest.Width == 0 {
		return
	}

	ellipsis := termbox.Cell{Ch: params.Ellipsis, Fg: params.Fg, Bg: params.Bg}
	off := dest.Y * this.Width + dest.X
	textlen := utf8.RuneCountInString(text)
	n := textlen
	if n > dest.Width {
		// string doesn't fit in the dest rectangle, draw ellipsis
		n = dest.Width - 1

		// if user asks for ellipsis in the center, alignment doesn't matter
		if params.CenterEllipsis {
			this.Cells[off+dest.Width/2] = ellipsis
		} else {
			switch params.Align {
			case AlignLeft:
				this.Cells[off+dest.Width-1] = ellipsis
			case AlignCenter:
				this.Cells[off] = ellipsis
				this.Cells[off+dest.Width-1] = ellipsis
				n--
			case AlignRight:
				this.Cells[off] = ellipsis
			}
		}
	}

	if n <= 0 {
		return
	}

	if params.CenterEllipsis && textlen != n {
		firsthalf := dest.Width / 2
		secondhalf := dest.Width - 1 - firsthalf
		this.draw_n_first_runes(off, firsthalf, params, text)
		off += dest.Width - 1
		this.draw_n_last_runes(off, secondhalf, params, text)
		return
	}

	switch params.Align {
	case AlignLeft:
		this.draw_n_first_runes(off, n, params, text)
	case AlignCenter:
		if textlen == n {
			off += (dest.Width - n) / 2
			this.draw_n_first_runes(off, n, params, text)
		} else {
			off++
			mid := (textlen - n) / 2
			text = skip_n_runes(text, mid)
			this.draw_n_first_runes(off, n, params, text)
		}
	case AlignRight:
		off += dest.Width - 1
		this.draw_n_last_runes(off, n, params, text)
	}
}