// Package icon 在运行时生成纯色圆形图标; 避免维护四套 ico 资源.
package icon

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/TarelX/AgentLamp/backend"
)

// Size 托盘图标边长 (像素); Windows 通知区习惯使用 16x16, 高 DPI 系统会自动放大
const Size = 16

// PNGForState 给定状态返回该状态对应的 PNG 字节; 颜色与前端灯泡一致
func PNGForState(s backend.AggregatedState) []byte {
	return drawCircle(colorForState(s))
}

func colorForState(s backend.AggregatedState) color.RGBA {
	switch s {
	case backend.StateIdle:
		return color.RGBA{0x22, 0xc5, 0x5e, 0xff} // green-500
	case backend.StateRunning:
		return color.RGBA{0xeA, 0xb3, 0x08, 0xff} // yellow-500
	case backend.StateWaiting:
		return color.RGBA{0xf5, 0x9e, 0x0b, 0xff} // amber-500
	case backend.StateError, backend.StateFault:
		return color.RGBA{0xef, 0x44, 0x44, 0xff} // red-500
	default:
		return color.RGBA{0x6b, 0x72, 0x80, 0xff} // gray-500
	}
}

// drawCircle 渲染一个填满 Size 的实心圆, 边缘做简单抗锯齿
func drawCircle(c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, Size, Size))
	cx, cy := float64(Size)/2-0.5, float64(Size)/2-0.5
	r := float64(Size)/2 - 1

	for y := 0; y < Size; y++ {
		for x := 0; x < Size; x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			dist := math.Sqrt(dx*dx + dy*dy)
			alpha := 0.0
			if dist <= r-0.5 {
				alpha = 1
			} else if dist <= r+0.5 {
				alpha = (r + 0.5 - dist)
			}
			if alpha > 0 {
				img.SetRGBA(x, y, color.RGBA{
					R: c.R, G: c.G, B: c.B,
					A: uint8(float64(c.A) * alpha),
				})
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
