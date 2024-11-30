//go:build windows && !nomedia

package win

import (
	"fmt"

	winrt "github.com/dece2183/media-winrt-go"
	"github.com/dece2183/media-winrt-go/windows/foundation"
	"github.com/go-ole/go-ole"
)

func awaitAsyncOperation(asyncOperation *foundation.IAsyncOperation, genericParamSignature string) error {
	waitChan := make(chan foundation.AsyncStatus, 2)

	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, genericParamSignature)
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), func(instance *foundation.AsyncOperationCompletedHandler, asyncInfo *foundation.IAsyncOperation, asyncStatus foundation.AsyncStatus) {
		waitChan <- asyncStatus
	})

	asyncOperation.SetCompleted(handler)
	status := <-waitChan
	handler.Release()
	close(waitChan)

	if status != foundation.AsyncStatusCompleted {
		return fmt.Errorf("async operation failed with status %d", status)
	}

	return nil
}
