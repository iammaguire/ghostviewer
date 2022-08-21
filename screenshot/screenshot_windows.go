package screenshot

import (
	"bytes"
	"fmt"
	"ghostviewer/d3d"
	"ghostviewer/win"
	"image"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

var (
	libUser32, _               = syscall.LoadLibrary("user32.dll")
	funcEnumDisplayMonitors, _ = syscall.GetProcAddress(syscall.Handle(libUser32), "EnumDisplayMonitors")
	ddapi                      = false
	DDUP                       *d3d.OutputDuplicator
	Device                     *d3d.ID3D11Device
	DeviceCtx                  *d3d.ID3D11DeviceContext
)

var ddApiEnabled []string = []string{
	"Windows 10",
	"Windows Server 2019",
	"Windows Server 2016",
	"Windows 8.1",
	"Windows Server 2012 R2",
	"Windows 8",
}

func InitializeScreenshotInterface() {
	if !HasDDAPI() {
		fmt.Println("No Desktop Duplication API on Windows 7 or earlier - using BitBlt, expect poor performance if Aero is enabled")
		return
	}

	if win.IsValidDpiAwarenessContext(win.DpiAwarenessContextPerMonitorAwareV2) {
		_, err := win.SetThreadDpiAwarenessContext(win.DpiAwarenessContextPerMonitorAwareV2)
		if err != nil {
			fmt.Printf("Could not set thread DPI awareness to PerMonitorAwareV2. %v\n", err)
		} else {
			fmt.Printf("Enabled PerMonitorAwareV2 DPI awareness.\n")
		}
	}

	device, deviceCtx, err := d3d.NewD3D11Device()
	Device = device
	DeviceCtx = deviceCtx
	if err != nil {
		fmt.Printf("Could not create D3D11 Device. %v\n", err)
		return
	}

	// TODO: This is just there, so that people can see how resizing might look
	//_ = resize.Resize(1920, 1080, imgBuf, resize.Bicubic)
	//_ = resize.Resize(1920, 1080, imgBuf, resize.Bicubic)

	fmt.Println("Desktop Duplication API present")
	ddapi = true
}

func HasDDAPI() bool {
	key, _ := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE) // error discarded for brevity
	defer key.Close()

	productName, _, err := key.GetStringValue("ProductName")
	if err != nil {
		return false
	}

	for _, v := range ddApiEnabled {
		if strings.Contains(v, productName) || strings.Contains(productName, v) {
			return true
		}
	}

	return false
}

func DDAPIScreenShot(bounds image.Rectangle) (*image.RGBA, error) {
	imgBuf := image.NewRGBA(bounds)
	if DDUP == nil {
		ddup, err := d3d.NewIDXGIOutputDuplication(Device, DeviceCtx, uint(0))
		DDUP = ddup
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}

	err := DDUP.GetImage(imgBuf, 0)
	if err != nil {
		return nil, err
	}

	//buf := bytes.Buffer{}
	//jpeg.Encode(bufio.NewWriter(&buf), imgBuf, nil)
	return imgBuf, nil
}

func ScreenRect() (image.Rectangle, error) {
	hDC := GetDC(0)
	if hDC == 0 {
		return image.Rectangle{}, fmt.Errorf("Could not Get primary display err:%d\n", GetLastError())
	}
	defer ReleaseDC(0, hDC)
	x := GetDeviceCaps(hDC, HORZRES)
	y := GetDeviceCaps(hDC, VERTRES)
	return image.Rect(0, 0, x, y), nil
}

func CaptureScreen() (*image.RGBA, error) {
	r, e := ScreenRect()
	if e != nil {
		return nil, e
	}
	if ddapi {
		return DDAPIScreenShot(r)
	} else {
		return CaptureRect(r)
	}
}

var fc = 0
var avg = 0
var total = 0

func CaptureRect(rect image.Rectangle) (*image.RGBA, error) {
	t := time.Now()

	hDC := GetDC(0)
	if hDC == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", GetLastError())
	}
	defer ReleaseDC(0, hDC)

	m_hDC := CreateCompatibleDC(hDC)
	if m_hDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", GetLastError())
	}
	defer DeleteDC(m_hDC)

	x, y := rect.Dx(), rect.Dy()

	bt := BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(x)
	bt.BmiHeader.BiHeight = int32(-y)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = BI_RGB

	ptr := unsafe.Pointer(uintptr(0))

	m_hBmp := CreateDIBSection(m_hDC, &bt, DIB_RGB_COLORS, &ptr, 0, 0)
	if m_hBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", GetLastError())
	}
	if m_hBmp == InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer DeleteObject(HGDIOBJ(m_hBmp))

	obj := SelectObject(m_hDC, HGDIOBJ(m_hBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", GetLastError())
	}
	if obj == 0xffffffff { //GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", GetLastError())
	}
	defer DeleteObject(obj)

	if !BitBlt(m_hDC, 0, 0, x, y, hDC, rect.Min.X, rect.Min.Y, SRCCOPY) {
		return nil, fmt.Errorf("BitBlt failed err:%d.\n", GetLastError())
	}

	var slice []byte
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hdrp.Data = uintptr(ptr)
	hdrp.Len = x * y * 4
	hdrp.Cap = x * y * 4
	imageBytes := make([]byte, len(slice))

	for i := 0; i < len(imageBytes); i += 4 {
		imageBytes[i], imageBytes[i+2], imageBytes[i+1], imageBytes[i+3] = slice[i+2], slice[i], slice[i+1], slice[i+3]
	}

	fc++
	ms := int(time.Now().Sub(t).Milliseconds())
	total += ms
	avg = total / fc

	avgFps := 1000.0 / avg
	fmt.Println(strconv.Itoa(avgFps) + " FPS")

	return &image.RGBA{imageBytes, 4 * x, image.Rect(0, 0, x, y)}, nil
}

func GetDeviceCaps(hdc HDC, index int) int {
	ret, _, _ := procGetDeviceCaps.Call(
		uintptr(hdc),
		uintptr(index))

	return int(ret)
}

func GetDC(hwnd HWND) HDC {
	ret, _, _ := procGetDC.Call(
		uintptr(hwnd))

	return HDC(ret)
}

func ReleaseDC(hwnd HWND, hDC HDC) bool {
	ret, _, _ := procReleaseDC.Call(
		uintptr(hwnd),
		uintptr(hDC))

	return ret != 0
}

func DeleteDC(hdc HDC) bool {
	ret, _, _ := procDeleteDC.Call(
		uintptr(hdc))

	return ret != 0
}

func GetLastError() uint32 {
	ret, _, _ := procGetLastError.Call()
	return uint32(ret)
}

func BitBlt(hdcDest HDC, nXDest, nYDest, nWidth, nHeight int, hdcSrc HDC, nXSrc, nYSrc int, dwRop uint) bool {
	ret, _, _ := procBitBlt.Call(
		uintptr(hdcDest),
		uintptr(nXDest),
		uintptr(nYDest),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(hdcSrc),
		uintptr(nXSrc),
		uintptr(nYSrc),
		uintptr(dwRop))

	return ret != 0
}

func SelectObject(hdc HDC, hgdiobj HGDIOBJ) HGDIOBJ {
	ret, _, _ := procSelectObject.Call(
		uintptr(hdc),
		uintptr(hgdiobj))

	if ret == 0 {
		panic("SelectObject failed")
	}

	return HGDIOBJ(ret)
}

func DeleteObject(hObject HGDIOBJ) bool {
	ret, _, _ := procDeleteObject.Call(
		uintptr(hObject))

	return ret != 0
}

func CreateDIBSection(hdc HDC, pbmi *BITMAPINFO, iUsage uint, ppvBits *unsafe.Pointer, hSection HANDLE, dwOffset uint) HBITMAP {
	ret, _, _ := procCreateDIBSection.Call(
		uintptr(hdc),
		uintptr(unsafe.Pointer(pbmi)),
		uintptr(iUsage),
		uintptr(unsafe.Pointer(ppvBits)),
		uintptr(hSection),
		uintptr(dwOffset))

	return HBITMAP(ret)
}

func CreateCompatibleDC(hdc HDC) HDC {
	ret, _, _ := procCreateCompatibleDC.Call(
		uintptr(hdc))

	if ret == 0 {
		panic("Create compatible DC failed")
	}

	return HDC(ret)
}

type (
	HANDLE  uintptr
	HWND    HANDLE
	HGDIOBJ HANDLE
	HDC     HANDLE
	HBITMAP HANDLE
)

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors *RGBQUAD
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type RGBQUAD struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

const (
	HORZRES          = 8
	VERTRES          = 10
	BI_RGB           = 0
	InvalidParameter = 2
	DIB_RGB_COLORS   = 0
	SRCCOPY          = 0x00CC0020
)

var (
	modgdi32               = syscall.NewLazyDLL("gdi32.dll")
	moduser32              = syscall.NewLazyDLL("user32.dll")
	modkernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetDC              = moduser32.NewProc("GetDC")
	procReleaseDC          = moduser32.NewProc("ReleaseDC")
	procDeleteDC           = modgdi32.NewProc("DeleteDC")
	procBitBlt             = modgdi32.NewProc("BitBlt")
	procDeleteObject       = modgdi32.NewProc("DeleteObject")
	procSelectObject       = modgdi32.NewProc("SelectObject")
	procCreateDIBSection   = modgdi32.NewProc("CreateDIBSection")
	procCreateCompatibleDC = modgdi32.NewProc("CreateCompatibleDC")
	procGetDeviceCaps      = modgdi32.NewProc("GetDeviceCaps")
	procGetLastError       = modkernel32.NewProc("GetLastError")
)

// Workaround for jpeg.Encode(), which requires a Flush()
// method to not call `bufio.NewWriter`
type bufferFlusher struct {
	bytes.Buffer
}

func (*bufferFlusher) Flush() error { return nil }
