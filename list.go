package main

import (
	"fmt"
	"image"

	. "github.com/gizak/termui/v3"
)

const (
	rowHeight = 3
)

type GaugeListItem struct {
	Percent int
	Label   string
}

func NewGaugeListItem(percent int, label string) *GaugeListItem {
	return &GaugeListItem{
		Percent: percent,
		Label:   label,
	}
}

type GaugeList struct {
	Block
	Rows             []*GaugeListItem
	SelectedRow      int
	SelectedRowStyle Style
	topRow           int
}

func NewGaugeList() *GaugeList {
	return &GaugeList{
		Block:            *NewBlock(),
		SelectedRowStyle: Theme.List.Text,
	}
}

func (self *GaugeList) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	// adjusts view into widget
	numRows := self.Inner.Dy() / rowHeight
	if self.SelectedRow > numRows-1 {
		self.topRow = self.SelectedRow - numRows + 1
	} else if self.SelectedRow < self.topRow {
		self.topRow = self.SelectedRow
	}

	// draw rows
	point := self.Inner.Min
	for row := self.topRow; row < len(self.Rows) && point.Y < self.Inner.Max.Y-2; row++ {
		gauge := self.Rows[row]

		// draw border
		self.drawBorder(buf, point, gauge.Label)

		// draw bar
		barWidth := int((float64(gauge.Percent) / 100) * float64(self.Inner.Dx()))
		c := ColorGreen
		if gauge.Percent >= 90 {
			c = ColorRed
		}
		buf.Fill(
			NewCell(' ', NewStyle(ColorClear, c)),
			image.Rect(point.X+2, point.Y+1, point.X+2+barWidth, point.Y+2),
		)

		// add percentage label
		label := fmt.Sprintf("%d%%", gauge.Percent)
		labelXCoordinate := point.X + 1 + (self.Inner.Dx() / 2) - int(float64(len(label))/2)
		labelYCoordinate := point.Y + 1
		if labelYCoordinate < self.Inner.Max.Y {
			for i, char := range label {
				style := NewStyle(ColorClear, ColorClear)
				if labelXCoordinate+i+1 <= point.X+barWidth {
					style = NewStyle(c, ColorClear, ModifierReverse)
				}
				buf.SetCell(NewCell(char, style), image.Pt(labelXCoordinate+i, labelYCoordinate))
			}
		}

		// add indicator if this is the selected row
		if row == self.SelectedRow {
			buf.SetCell(NewCell('*', NewStyle(ColorClear, ColorClear, ModifierBold)), image.Pt(point.X, point.Y+1))
		}

		// update the starting point for the next row
		point = image.Pt(self.Inner.Min.X, point.Y+rowHeight)
	}

	// draw UP_ARROW if needed
	if self.topRow > 0 {
		buf.SetCell(
			NewCell(UP_ARROW, NewStyle(ColorWhite)),
			image.Pt(self.Inner.Max.X-1, self.Inner.Min.Y),
		)
	}

	// draw DOWN_ARROW if needed
	if self.topRow+numRows < len(self.Rows) {
		buf.SetCell(
			NewCell(DOWN_ARROW, NewStyle(ColorWhite)),
			image.Pt(self.Inner.Max.X-1, self.Inner.Max.Y-1),
		)
	}
}

func (self *GaugeList) drawBorder(buf *Buffer, point image.Point, label string) {
	verticalCell := NewCell(VERTICAL_LINE, NewStyle(ColorClear, ColorClear))
	horizontalCell := NewCell(HORIZONTAL_LINE, NewStyle(ColorClear, ColorClear))
	buf.Fill(horizontalCell, image.Rect(point.X+1, point.Y, self.Inner.Max.X, point.Y+1))
	buf.Fill(horizontalCell, image.Rect(point.X+1, point.Y+2, self.Inner.Max.X, point.Y+3))
	buf.Fill(verticalCell, image.Rect(point.X+2, point.Y+1, point.X+1, point.Y+2))
	buf.Fill(verticalCell, image.Rect(self.Inner.Max.X-1, point.Y+1, self.Inner.Max.X, point.Y+2))
	buf.SetCell(NewCell(TOP_LEFT, NewStyle(ColorClear, ColorClear)), image.Pt(point.X+1, point.Y))
	buf.SetCell(NewCell(TOP_RIGHT, NewStyle(ColorClear, ColorClear)), image.Pt(self.Inner.Max.X-1, point.Y))
	buf.SetCell(NewCell(BOTTOM_LEFT, NewStyle(ColorClear, ColorClear)), image.Pt(point.X+1, point.Y+2))
	buf.SetCell(NewCell(BOTTOM_RIGHT, NewStyle(ColorClear, ColorClear)), image.Pt(self.Inner.Max.X-1, point.Y+2))
	buf.SetString(
		" "+label+" ",
		NewStyle(ColorClear),
		image.Pt(point.X+2, point.Y),
	)
}

// ScrollAmount scrolls by amount given. If amount is < 0, then scroll up.
// There is no need to set self.topRow, as this will be set automatically when drawn,
// since if the selected item is off screen then the topRow variable will change accordingly.
func (self *GaugeList) ScrollAmount(amount int) {
	if len(self.Rows)-int(self.SelectedRow) <= amount {
		self.SelectedRow = len(self.Rows) - 1
	} else if int(self.SelectedRow)+amount < 0 {
		self.SelectedRow = 0
	} else {
		self.SelectedRow += amount
	}
}

func (self *GaugeList) ScrollUp() {
	self.ScrollAmount(-1)
}

func (self *GaugeList) ScrollDown() {
	self.ScrollAmount(1)
}

func (self *GaugeList) ScrollPageUp() {
	// If an item is selected below top row, then go to the top row.
	if self.SelectedRow > self.topRow {
		self.SelectedRow = self.topRow
	} else {
		self.ScrollAmount(-self.Inner.Dy())
	}
}

func (self *GaugeList) ScrollPageDown() {
	self.ScrollAmount(self.Inner.Dy())
}

func (self *GaugeList) ScrollHalfPageUp() {
	self.ScrollAmount(-int(FloorFloat64(float64(self.Inner.Dy()) / 2)))
}

func (self *GaugeList) ScrollHalfPageDown() {
	self.ScrollAmount(int(FloorFloat64(float64(self.Inner.Dy()) / 2)))
}

func (self *GaugeList) ScrollTop() {
	self.SelectedRow = 0
}

func (self *GaugeList) ScrollBottom() {
	self.SelectedRow = len(self.Rows) - 1
}
