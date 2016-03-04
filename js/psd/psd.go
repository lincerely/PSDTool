//go:generate gopherjs build -m

package main

import (
	"archive/zip"
	"crypto/md5"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/oov/psd"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/japanese"
)

type root struct {
	X        int
	Y        int
	Width    int
	Height   int
	Children []layer

	CanvasWidth  int
	CanvasHeight int
	Hash         string
	PFV          string
	Readme       string

	processed int
	progress  func(l *layer)
	realRect  image.Rectangle
}

type layer struct {
	X        int
	Y        int
	Width    int
	Height   int
	Children []layer

	Name                  string
	BlendMode             string
	Opacity               uint8
	Clipping              bool
	BlendClippedElements  bool
	TransparencyProtected bool
	Visible               bool
	Canvas                *js.Object
	MaskX                 int
	MaskY                 int
	MaskCanvas            *js.Object
	MaskDefaultColor      int
	Folder                bool
	FolderOpen            bool
	psdLayer              *psd.Layer
}

func main() {
	// psd.Debug = log.New(os.Stdout, "psd: ", log.Lshortfile)
	js.Global.Set("parsePSD", parsePSD)
}

func arrayBufferToByteSlice(a *js.Object) []byte {
	return js.Global.Get("Uint8Array").New(a).Interface().([]byte)
}

func (r *root) buildLayer(l *layer) error {
	var err error

	if l.psdLayer.UnicodeName == "" && l.psdLayer.MBCSName != "" {
		if l.Name, err = japanese.ShiftJIS.NewDecoder().String(l.psdLayer.MBCSName); err != nil {
			l.Name = l.psdLayer.MBCSName
		}
	} else {
		l.Name = l.psdLayer.UnicodeName
	}
	if l.psdLayer.Folder() {
		l.BlendMode = l.psdLayer.SectionDividerSetting.BlendMode.String()
	} else {
		l.BlendMode = l.psdLayer.BlendMode.String()
	}
	l.Opacity = l.psdLayer.Opacity
	l.Clipping = l.psdLayer.Clipping
	l.BlendClippedElements = l.psdLayer.BlendClippedElements
	l.Visible = l.psdLayer.Visible()
	l.Folder = l.psdLayer.Folder()
	l.FolderOpen = l.psdLayer.FolderIsOpen()

	if l.psdLayer.HasImage() && l.psdLayer.Rect.Dx()*l.psdLayer.Rect.Dy() > 0 {
		if l.Canvas, err = createImageCanvas(l.psdLayer); err != nil {
			return err
		}
		r.realRect = r.realRect.Union(l.psdLayer.Rect)
	}
	if _, ok := l.psdLayer.Channel[-2]; ok && l.psdLayer.Mask.Enabled() && l.psdLayer.Mask.Rect.Dx()*l.psdLayer.Mask.Rect.Dy() > 0 {
		if l.MaskCanvas, err = createMaskCanvas(l.psdLayer); err != nil {
			return err
		}
		l.MaskX = l.psdLayer.Mask.Rect.Min.X
		l.MaskY = l.psdLayer.Mask.Rect.Min.Y
		l.MaskDefaultColor = l.psdLayer.Mask.DefaultColor
	}

	r.processed++
	r.progress(l)

	rect := l.psdLayer.Rect
	for i := range l.psdLayer.Layer {
		l.Children = append(l.Children, layer{psdLayer: &l.psdLayer.Layer[i]})
		if err = r.buildLayer(&l.Children[i]); err != nil {
			return err
		}
		rect = rect.Union(image.Rect(
			l.Children[i].X,
			l.Children[i].Y,
			l.Children[i].X+l.Children[i].Width,
			l.Children[i].Y+l.Children[i].Height,
		))
	}
	l.X = rect.Min.X
	l.Y = rect.Min.Y
	l.Width = rect.Dx()
	l.Height = rect.Dy()
	return nil
}

func countLayers(l []psd.Layer) int {
	r := len(l)
	for i := range l {
		r += countLayers(l[i].Layer)
	}
	return r
}

func (r *root) Build(img *psd.PSD, progress func(processed, total int, l *layer)) error {
	numLayers := countLayers(img.Layer)
	r.CanvasWidth = img.Config.Rect.Dx()
	r.CanvasHeight = img.Config.Rect.Dy()
	r.progress = func(l *layer) { progress(r.processed, numLayers, l) }
	for i := range img.Layer {
		r.Children = append(r.Children, layer{psdLayer: &img.Layer[i]})
		if err := r.buildLayer(&r.Children[i]); err != nil {
			return err
		}
	}
	r.realRect = r.realRect.Intersect(image.Rect(0, 0, r.CanvasWidth, r.CanvasHeight))
	r.X = r.realRect.Min.X
	r.Y = r.realRect.Min.Y
	r.Width = r.realRect.Dx()
	r.Height = r.realRect.Dy()
	return nil
}

func createImageCanvas(l *psd.Layer) (*js.Object, error) {
	if l.Picker.ColorModel() != color.NRGBAModel {
		return nil, errors.New("Unsupported color mode")
	}

	sw, sh := l.Rect.Dx(), l.Rect.Dy()
	cvs := createCanvas(sw, sh)
	ctx := cvs.Call("getContext", "2d")
	imgData := ctx.Call("createImageData", sw, sh)
	dw := imgData.Get("width").Int()
	data := imgData.Get("data")

	var ofsd, ofss, x, y, sx, dx int
	r, g, b := l.Channel[0], l.Channel[1], l.Channel[2]
	rp, gp, bp := r.Data, g.Data, b.Data
	if a, ok := l.Channel[-1]; ok {
		ap := a.Data
		for y = 0; y < sh; y++ {
			ofss, ofsd = y*sw, y*dw<<2
			for x = 0; x < sw; x++ {
				sx, dx = ofss+x, ofsd+x<<2
				data.SetIndex(dx+0, rp[sx])
				data.SetIndex(dx+1, gp[sx])
				data.SetIndex(dx+2, bp[sx])
				data.SetIndex(dx+3, ap[sx])
			}
		}
	} else {
		for y = 0; y < sh; y++ {
			ofss, ofsd = y*sw, y*dw<<2
			for x = 0; x < sw; x++ {
				sx, dx = ofss+x, ofsd+x<<2
				data.SetIndex(dx+0, rp[sx])
				data.SetIndex(dx+1, gp[sx])
				data.SetIndex(dx+2, bp[sx])
				data.SetIndex(dx+3, 0xff)
			}
		}
	}
	ctx.Call("putImageData", imgData, 0, 0)
	return cvs, nil
}

func createMaskCanvas(l *psd.Layer) (*js.Object, error) {
	m, ok := l.Channel[-2]
	if !ok {
		return nil, nil
	}

	sw, sh := l.Mask.Rect.Dx(), l.Mask.Rect.Dy()
	cvs := createCanvas(sw, sh)
	ctx := cvs.Call("getContext", "2d")
	imgData := ctx.Call("createImageData", sw, sh)
	dw := imgData.Get("width").Int()
	data := imgData.Get("data")

	var ofsd, ofss, x, y, sx, dx int
	mp := m.Data
	if l.Mask.DefaultColor == 0 {
		for y = 0; y < sh; y++ {
			ofss, ofsd = y*sw, y*dw<<2
			for x = 0; x < sw; x++ {
				sx, dx = ofss+x, ofsd+x<<2
				data.SetIndex(dx+3, mp[sx])
			}
		}
	} else {
		for y = 0; y < sh; y++ {
			ofss, ofsd = y*sw, y*dw<<2
			for x = 0; x < sw; x++ {
				sx, dx = ofss+x, ofsd+x<<2
				data.SetIndex(dx+3, 255-mp[sx])
			}
		}
	}
	ctx.Call("putImageData", imgData, 0, 0)
	return cvs, nil
}

func createCanvas(width, height int) *js.Object {
	cvs := js.Global.Get("document").Call("createElement", "canvas")
	cvs.Set("width", width)
	cvs.Set("height", height)
	return cvs
}

func readTextFile(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	d, err := chardet.NewTextDetector().DetectBest(b)
	if err != nil {
		return "", err
	}

	switch d.Charset {
	case "ISO-2022-JP":
		b, err = japanese.ISO2022JP.NewDecoder().Bytes(b)
	case "EUC-JP":
		b, err = japanese.EUCJP.NewDecoder().Bytes(b)
	case "Shift_JIS":
		b, err = japanese.ShiftJIS.NewDecoder().Bytes(b)
	case "UTF-8":
		break
	default:
		return "", errors.New("unsupported charset: " + d.Charset)
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type reader interface {
	io.Reader
	Sum() []byte
}

func parse(b []byte, progress func(step string, progress float64, l *layer)) (*root, error) {
	var r root
	s := time.Now().UnixNano()
	if len(b) < 4 {
		return nil, errors.New("unsupported file type")
	}
	var reader reader
	switch string(b[:4]) {
	case "PK\x03\x04": // zip archive
		zr, err := zip.NewReader(&progressReader{Buf: b}, int64(len(b)))
		if err != nil {
			return nil, err
		}
		var psdf, pfvf, txtf *zip.File
		for _, f := range zr.File {
			if psdf == nil && strings.ToLower(f.Name[len(f.Name)-4:]) == ".psd" {
				psdf = f
				continue
			}
			if pfvf == nil && strings.ToLower(f.Name[len(f.Name)-4:]) == ".pfv" {
				pfvf = f
				continue
			}
			if txtf == nil && strings.ToLower(f.Name[len(f.Name)-4:]) == ".txt" {
				txtf = f
				continue
			}
		}
		if psdf == nil {
			return nil, errors.New("psd file is not found from given zip archive")
		}

		if pfvf != nil {
			pfvr, err := pfvf.Open()
			if err != nil {
				return nil, err
			}
			defer pfvr.Close()
			r.PFV, err = readTextFile(pfvr)
			if err != nil {
				return nil, err
			}
		}

		if txtf != nil {
			txtr, err := txtf.Open()
			if err != nil {
				return nil, err
			}
			defer txtr.Close()
			r.Readme, err = readTextFile(txtr)
			if err != nil {
				return nil, err
			}
		}

		rc, err := psdf.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		reader = &genericProgressReader{
			R:        rc,
			Hash:     md5.New(),
			Progress: func(p float64) { progress("parse", p, nil) },
			ln:       int64(psdf.UncompressedSize64),
		}
	case "7z\xbc\xaf": // 7z archive
		return nil, errors.New("7z archive is not supported")
	case "8BPS": // psd file
		reader = &progressReader{
			Buf:      b,
			Hash:     md5.New(),
			Progress: func(p float64) { progress("parse", p, nil) },
		}
		break
	default:
		return nil, errors.New("unsupported file type")
	}
	psdImg, _, err := psd.Decode(reader, nil)
	if err != nil {
		return nil, err
	}
	e := time.Now().UnixNano()
	progress("parse", 1, nil)
	log.Println("Decode PSD Structure:", (e-s)/1e6)

	if psdImg.Config.ColorMode != psd.ColorModeRGB {
		return nil, errors.New("Unsupported color mode")
	}

	s = time.Now().UnixNano()
	r.Hash = fmt.Sprintf("%x", reader.Sum())
	if err = r.Build(psdImg, func(processed, total int, l *layer) {
		progress("draw", float64(processed)/float64(total), l)
	}); err != nil {
		return nil, err
	}
	e = time.Now().UnixNano()
	log.Println("Build Canvas:", (e-s)/1e6)
	return &r, nil
}

func parsePSD(in *js.Object, progress *js.Object, complete *js.Object, failed *js.Object) {
	go func() {
		next := time.Now()
		root, err := parse(arrayBufferToByteSlice(in), func(step string, prog float64, l *layer) {
			if now := time.Now(); now.After(next) {
				progress.Invoke(step, prog, l)
				time.Sleep(1) // anti-freeze
				next = now.Add(100 * time.Millisecond)
			}
		})
		if err != nil {
			failed.Invoke(err.Error())
			return
		}
		complete.Invoke(root)
	}()
}
