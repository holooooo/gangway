package router

import (
	"io"
)

func handleARP(writer io.Writer, pkg []byte) {
	res := make([]byte, 0)
	res = append(res, pkg[8:14]...)
	res = append(res, pkg[8:14]...)
	res = append(res, 8, 6)

	res = append(res, 0, 1, 8, 0, 6, 4, 0, 2)
	res = append(res, pkg[8:14]...)
	res = append(res, pkg[24:28]...)
	res = append(res, pkg[8:14]...)
	res = append(res, pkg[14:18]...)

	_, _ = writer.Write(res)
}
