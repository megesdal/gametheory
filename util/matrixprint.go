package util

import (
	"fmt"
	"io"
)

type MatrixPrinter struct {
	buf [][]string

	ncols     int
	colwidths []int
	colaligns []int

	col  int
	line []string

	value string
}

func NewMatrixPrinter() *MatrixPrinter {
	return &MatrixPrinter{
		buf:       make([][]string, 0),
		ncols:     0,
		colwidths: make([]int, 0),
		colaligns: make([]int, 0),
		col:       0,
		line:      make([]string, 0),
	}
}

func (out *MatrixPrinter) Colipr(i int) {
	out.Colpr(fmt.Sprintf("%d", i))
}

func (out *MatrixPrinter) Colpr(s string) *MatrixPrinter {

	swidth := len(s)
	if out.col == len(out.colwidths) {
		out.colwidths = append(out.colwidths, swidth)
	} else if out.colwidths[out.col] < swidth {
		out.colwidths[out.col] = swidth
	}

	if out.col == len(out.colaligns) {
		out.colaligns = append(out.colaligns, 1)
	}

	if out.col == len(out.line) {
		out.line = append(out.line, s)
	} else {
		out.line[out.col] = s
	}

	out.col += 1
	return out
}

/* making next column written to left-adjusted      */
func (out *MatrixPrinter) Colleft() *MatrixPrinter {

	if out.col == len(out.colaligns) {
		out.colaligns = append(out.colaligns, -1)
	} else {
		out.colaligns[out.col] = -1
	}
	return out
}

func (out *MatrixPrinter) Colnl() *MatrixPrinter {

	out.buf = append(out.buf, out.line)
	out.col = 0
	out.line = make([]string, len(out.colwidths))
	return out
}

func (out *MatrixPrinter) Colout(w io.Writer) {

	for i := 0; i < len(out.buf); i++ {
		line := out.buf[i]
		out.printRow(w, line)
		io.WriteString(w, "\n")
	}
	out.printRow(w, out.line)
	io.WriteString(w, "\n") //Do we want a newline here?
}

func (out *MatrixPrinter) printRow(w io.Writer, line []string) {

	first := false
	for j := 0; j < len(line); j++ {
		width := out.colwidths[j] * out.colaligns[j]
		if width != 0 {
			if first {
				first = false
			} else {
				io.WriteString(w, " ")
			}
			format := fmt.Sprintf("%%%ds", width)
			io.WriteString(w, fmt.Sprintf(format, line[j]))
		}
	}
}

/*
func (out *MatrixPrinter) sortBy(final int col, final int fromRow) {

    	List<List<String>> subBuf = buf.subList(fromRow, buf.size());
    	buf = buf.subList(0, fromRow);
    	Collections.sort(subBuf, new Comparator<List<String>>() {

			public int compare(List<String> a, List<String> b) {
				String aStr = a.get(col);
				String bStr = b.get(col);
				try {
					return Integer.valueOf(aStr).compareTo(Integer.valueOf(bStr));
				} catch (NumberFormatException ex) {
					return aStr.compareTo(bStr);
				}
			}

    	});
    	buf.addAll(subBuf);
    }
}
*/
