/*
Copyright © 2020 Chris Kim

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package widgets

import (
	"bytes"
	"fmt"
	"image"
	"strings"
	"text/tabwriter"

	. "github.com/gizak/termui/v3"
	rw "github.com/mattn/go-runewidth"

	"github.com/chriskim06/kubectl-ptop/internal/metrics"
)

type tabbedLine struct {
	content string
}

// ResourceList is a custom widget that displays normal kubectl top output
// and color indicators to match up with a KubePlot
type ResourceList struct {
	Block

	Headers          []string
	Metrics          []metrics.MetricsValues
	TextStyle        Style
	SelectedRow      int
	topRow           int
	SelectedRowStyle Style
}

// NewResourceList instantiates a new resource list widget
func NewResourceList() *ResourceList {
	return &ResourceList{
		Block:            *NewBlock(),
		TextStyle:        Theme.List.Text,
		SelectedRowStyle: Theme.List.Text,
	}
}

// Draw renders the resource list
func (self *ResourceList) Draw(buf *Buffer) {
	lines := getTabbedStringList(self.Inner.Dx(), self.Headers, self.Metrics)
	self.Title = lines[0].content
	lines = lines[1:]

	self.Block.Draw(buf)

	// adjusts view into widget
	if self.SelectedRow >= self.Inner.Dy()+self.topRow {
		self.topRow = self.SelectedRow - self.Inner.Dy() + 1
	} else if self.SelectedRow < self.topRow {
		self.topRow = self.SelectedRow
	}

	point := self.Inner.Min

	// draw lines
	for row := self.topRow; row < len(lines) && point.Y < self.Inner.Max.Y; row++ {
		// draw the color indicator first
		indicatorStyle := NewStyle(ColorClear, ColorClear)
		if row == self.SelectedRow {
			indicatorStyle.Modifier = ModifierBold
			buf.SetCell(NewCell('*', indicatorStyle), point)
		}
		point = point.Add(image.Pt(2, 0))

		// draw the content for the line next
		line := lines[row].content
		for _, char := range line {
			if point.Y >= self.Inner.Max.Y {
				break
			}

			style := NewStyle(ColorClear, ColorClear)
			if row == self.SelectedRow {
				style = self.SelectedRowStyle
			}

			if point.X+1 == self.Inner.Max.X+1 && len(line) > self.Inner.Dx() {
				buf.SetCell(NewCell(ELLIPSES, style), point.Add(image.Pt(-1, 0)))
				break
			} else {
				buf.SetCell(NewCell(char, style), point)
				point = point.Add(image.Pt(rw.RuneWidth(char), 0))
			}
		}
		point = image.Pt(self.Inner.Min.X, point.Y+1)
	}

	// draw UP_ARROW if needed
	if self.topRow > 0 {
		buf.SetCell(
			NewCell(UP_ARROW, NewStyle(ColorClear)),
			image.Pt(self.Inner.Max.X-1, self.Inner.Min.Y),
		)
	}

	// draw DOWN_ARROW if needed
	if len(self.Metrics) > int(self.topRow)+self.Inner.Dy() {
		buf.SetCell(
			NewCell(DOWN_ARROW, NewStyle(ColorClear)),
			image.Pt(self.Inner.Max.X-1, self.Inner.Max.Y-1),
		)
	}
}

// ScrollAmount scrolls by amount given. If amount is < 0, then scroll up.
// There is no need to set self.topRow, as this will be set automatically when drawn,
// since if the selected item is off screen then the topRow variable will change accordingly.
func (self *ResourceList) ScrollAmount(amount int) {
	if len(self.Metrics)-int(self.SelectedRow) <= amount {
		self.SelectedRow = len(self.Metrics) - 1
	} else if int(self.SelectedRow)+amount < 0 {
		self.SelectedRow = 0
	} else {
		self.SelectedRow += amount
	}
}

// ScrollUp scrolls up by 1
func (self *ResourceList) ScrollUp() {
	self.ScrollAmount(-1)
}

// ScrollDown scrolls down by 1
func (self *ResourceList) ScrollDown() {
	self.ScrollAmount(1)
}

// ScrollTop scrolls to the top of the list
func (self *ResourceList) ScrollTop() {
	self.SelectedRow = 0
}

// ScrollBottom scrolls to the bottom of the list
func (self *ResourceList) ScrollBottom() {
	self.SelectedRow = len(self.Metrics) - 1
}

func getTabbedStringList(width int, headers []string, metricsValues []metrics.MetricsValues) []tabbedLine {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)

	// add the column names
	for _, header := range headers {
		fmt.Fprintf(w, "%s\t", header)
	}
	fmt.Fprint(w, " \n")

	// add the metrics themselves
	half := width / 2
	for i, m := range metricsValues {
		name := m.Name
		if len(name) > half {
			name = fmt.Sprintf("%s%c", name[0:half-3], ELLIPSES)
		}
		fmt.Fprintf(w, "%s\t%dm\t%.2f\t%dMi\t%.2f\t ", name, m.CPUCores, m.CPUPercent, m.MemCores, m.MemPercent)
		if i != len(metricsValues)-1 {
			fmt.Fprint(w, "\n")
		}
	}
	w.Flush()
	lines := strings.Split(b.String(), "\n")

	tabbedLines := []tabbedLine{}
	for i, line := range lines {
		// this is the column names line
		if i == 0 {
			tabbedLines = append(tabbedLines, tabbedLine{
				content: line,
			})
			continue
		}

		tabbedLines = append(tabbedLines, tabbedLine{
			content: line,
		})
	}

	return tabbedLines
}
