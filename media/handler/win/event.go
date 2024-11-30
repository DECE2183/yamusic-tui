//go:build windows && !nomedia

package win

import (
	winrt "github.com/dece2183/media-winrt-go"
	"github.com/dece2183/media-winrt-go/windows/foundation"
	"github.com/go-ole/go-ole"
)

func makeEventHandler(callback foundation.TypedEventHandlerCallback, paramSignatures ...string) *foundation.TypedEventHandler {
	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDTypedEventHandler, paramSignatures...)
	handler := foundation.NewTypedEventHandler(ole.NewGUID(iid), callback)
	return handler
}
