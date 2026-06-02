package icon

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// LampMaskPNG 生成 width × height 的 PNG 蒙版.
// 不透明区域 = 灯壳 + 顶部安装柱; 其他像素 alpha=0 让窗口完全不渲染,
// 主屏 DWM aero shadow 与副屏 DPI 切换的白底退化都被裁掉.
func LampMaskPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{255, 255, 255, 255}

	// 灯壳: 110×252 圆角矩形, mask 外完全不画
	bodyW, bodyH := 110, 252
	bodyX := (width - bodyW) / 2
	bodyY := (height - bodyH) / 2
	const bodyRadius = 24
	drawRoundRect(img, bodyX, bodyY, bodyX+bodyW, bodyY+bodyH, bodyRadius, white)

	// 顶部安装柱: 16×30, 紧贴主体上方居中
	stalkW, stalkH := 16, 30
	stalkX := (width - stalkW) / 2
	stalkY := bodyY - stalkH + 2
	drawRoundRect(img, stalkX, stalkY, stalkX+stalkW, stalkY+stalkH, 4, white)

	// 底部铭牌区: 包住 ::after 文字, 让 CURSOR / AGENTLAMP 可见
	plateW, plateH := 96, 18
	plateX := (width - plateW) / 2
	plateY := bodyY + bodyH - 2
	drawRoundRect(img, plateX, plateY, plateX+plateW, plateY+plateH, 6, white)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// drawRoundRect 在 [x0,y0,x1,y1) 内绘制圆角实心矩形, r 为圆角半径
func drawRoundRect(img *image.RGBA, x0, y0, x1, y1, r int, c color.RGBA) {
	if r < 0 {
		r = 0
	}
	if r*2 > x1-x0 {
		r = (x1 - x0) / 2
	}
	if r*2 > y1-y0 {
		r = (y1 - y0) / 2
	}
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if pointInRoundRect(x, y, x0, y0, x1, y1, r) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func pointInRoundRect(x, y, x0, y0, x1, y1, r int) bool {
	if x < x0 || x >= x1 || y < y0 || y >= y1 {
		return false
	}
	// 中心带状区域必在内
	if x >= x0+r && x < x1-r {
		return true
	}
	if y >= y0+r && y < y1-r {
		return true
	}
	// 四角圆形测试
	corners := [4][2]int{
		{x0 + r, y0 + r},
		{x1 - 1 - r, y0 + r},
		{x0 + r, y1 - 1 - r},
		{x1 - 1 - r, y1 - 1 - r},
	}
	for _, c := range corners {
		dx, dy := x-c[0], y-c[1]
		if dx*dx+dy*dy <= r*r {
			return true
		}
	}
	return false
}
