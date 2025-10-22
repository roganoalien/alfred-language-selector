//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Carbon
#include <Carbon/Carbon.h>

static TISInputSourceRef getCurrentSource() {
    return TISCopyCurrentKeyboardInputSource();
}

static CFArrayRef getAllSources() {
    CFDictionaryRef props = CFDictionaryCreate(
        NULL,
        (const void *[]){ kTISPropertyInputSourceCategory },
        (const void *[]){ kTISCategoryKeyboardInputSource },
        1,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );
    return TISCreateInputSourceList(props, false);
}

static CFStringRef getSourceName(TISInputSourceRef src) {
    return TISGetInputSourceProperty(src, kTISPropertyLocalizedName);
}

static CFStringRef getSourceID(TISInputSourceRef src) {
    return TISGetInputSourceProperty(src, kTISPropertyInputSourceID);
}

static void selectSource(CFStringRef sid) {
    CFDictionaryRef props = CFDictionaryCreate(
        NULL,
        (const void *[]){ kTISPropertyInputSourceID },
        (const void *[]){ sid },
        1,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );
    CFArrayRef list = TISCreateInputSourceList(props, false);
    if (CFArrayGetCount(list) > 0) {
        TISInputSourceRef src = (TISInputSourceRef)CFArrayGetValueAtIndex(list, 0);
        TISSelectInputSource(src);
    }
}
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"os"
	"unsafe"
)

type AlfredItem struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Arg      string `json:"arg"`
	Valid    bool   `json:"valid"`
}

type AlfredOutput struct {
	Items []AlfredItem `json:"items"`
}

func cfStringToGo(s C.CFStringRef) string {
	if s == 0 {
		return ""
	}
	var buffer [1024]C.char
	C.CFStringGetCString(s, &buffer[0], 1024, C.kCFStringEncodingUTF8)
	return C.GoString(&buffer[0])
}

func getInputSources(shouldReturn bool) (AlfredOutput, int) {
	current := C.getCurrentSource()
	// currentName := cfStringToGo(C.getSourceName(current))
	currentID := cfStringToGo(C.getSourceID(current))

	all := C.getAllSources()
	count := int(C.CFArrayGetCount(all))

	var items []AlfredItem
	var currentIndex int

	for i := range count {
		src := (C.TISInputSourceRef)(C.CFArrayGetValueAtIndex(all, C.CFIndex(i)))
		name := cfStringToGo(C.getSourceName(src))
		id := cfStringToGo(C.getSourceID(src))
		mark := ""
		if id == currentID {
			currentIndex = i
			mark = " (Current)"
		}
		item := AlfredItem{
			Title:    name + mark,
			Subtitle: id,
			Arg:      id,
			Valid:    true,
		}
		items = append(items, item)
	}

	out := AlfredOutput{Items: items}

	if shouldReturn {
		return out, currentIndex
	} else {
		json.NewEncoder(os.Stdout).Encode(out)
		return AlfredOutput{}, 0
	}
}

func selectIndex(next bool) {
	out, index := getInputSources(true)
	if next {
		switchLayout(out.Items[index+1].Arg)
	} else {
		switchLayout(out.Items[index-1].Arg)
	}
}

func listLayouts() {
	getInputSources(false)
}

func switchLayout(idStr string) {
	cstr := C.CString(idStr)
	defer C.free(unsafe.Pointer(cstr))

	id := C.CFStringCreateWithCString(C.kCFAllocatorDefault, cstr, C.kCFStringEncodingUTF8)
	C.selectSource(id)
	fmt.Printf("Switched to: %s\n", idStr)
}

func main() {
	args := os.Args

	if len(args) < 2 {
		listLayouts()
		return
	}

	param := args[1]

	switch param {
	case "next", "prev":
		selectIndex(param == "next")
	default:
		switchLayout(param)
	}
}
