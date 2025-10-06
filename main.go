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

func listLayouts() {
	current := C.getCurrentSource()
	// currentName := cfStringToGo(C.getSourceName(current))
	currentID := cfStringToGo(C.getSourceID(current))

	all := C.getAllSources()
	count := int(C.CFArrayGetCount(all))

	var items []AlfredItem

	for i := 0; i < count; i++ {
		src := (C.TISInputSourceRef)(C.CFArrayGetValueAtIndex(all, C.CFIndex(i)))
		name := cfStringToGo(C.getSourceName(src))
		id := cfStringToGo(C.getSourceID(src))
		mark := ""
		if id == currentID {
			mark = " (Current)"
		}
		items = append(items, AlfredItem{
			Title:    name + mark,
			Subtitle: id,
			Arg:      id,
			Valid:    true,
		})
	}

	out := AlfredOutput{Items: items}
	json.NewEncoder(os.Stdout).Encode(out)
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
	if len(args) == 1 {
		listLayouts()
	} else if len(args) == 2 {
		switchLayout(args[1])
	} else {
		fmt.Println("Usage:")
		fmt.Println("  keyboardlang           # list layouts as JSON for Alfred")
		fmt.Println("  keyboardlang <sourceID> # switch layout")
	}
}
