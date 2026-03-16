//go:build darwin && !nomedia

package macos

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework MediaPlayer -framework IOKit

#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <MediaPlayer/MediaPlayer.h>
#import <IOKit/hidsystem/ev_keymap.h>
#include <pthread.h>

// Go callbacks declared here, implemented in Go via //export
extern void goOnPlay(void);
extern void goOnPause(void);
extern void goOnTogglePlayPause(void);
extern void goOnNext(void);
extern void goOnPrevious(void);
extern void goOnSeek(double positionSeconds);
extern void goOnSetVolume(double volume);
extern void goOnStop(void);

static void setupRemoteCommandCenter(void) {
    MPRemoteCommandCenter *center = [MPRemoteCommandCenter sharedCommandCenter];

    [center.playCommand setEnabled:YES];
    [center.playCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goOnPlay();
        return MPRemoteCommandHandlerStatusSuccess;
    }];

    [center.pauseCommand setEnabled:YES];
    [center.pauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goOnPause();
        return MPRemoteCommandHandlerStatusSuccess;
    }];

    [center.togglePlayPauseCommand setEnabled:YES];
    [center.togglePlayPauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goOnTogglePlayPause();
        return MPRemoteCommandHandlerStatusSuccess;
    }];

    [center.nextTrackCommand setEnabled:YES];
    [center.nextTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goOnNext();
        return MPRemoteCommandHandlerStatusSuccess;
    }];

    [center.previousTrackCommand setEnabled:YES];
    [center.previousTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goOnPrevious();
        return MPRemoteCommandHandlerStatusSuccess;
    }];

    [center.changePlaybackPositionCommand setEnabled:YES];
    [center.changePlaybackPositionCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        MPChangePlaybackPositionCommandEvent *posEvent = (MPChangePlaybackPositionCommandEvent *)event;
        goOnSeek(posEvent.positionTime);
        return MPRemoteCommandHandlerStatusSuccess;
    }];

    [center.stopCommand setEnabled:YES];
    [center.stopCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        goOnStop();
        return MPRemoteCommandHandlerStatusSuccess;
    }];
}

static void teardownRemoteCommandCenter(void) {
    MPRemoteCommandCenter *center = [MPRemoteCommandCenter sharedCommandCenter];
    [center.playCommand removeTarget:nil];
    [center.pauseCommand removeTarget:nil];
    [center.togglePlayPauseCommand removeTarget:nil];
    [center.nextTrackCommand removeTarget:nil];
    [center.previousTrackCommand removeTarget:nil];
    [center.changePlaybackPositionCommand removeTarget:nil];
    [center.stopCommand removeTarget:nil];
}

static void runLoop(void) {
    NSRunLoop *loop = [NSRunLoop currentRunLoop];
    while (YES) {
        [loop runMode:NSDefaultRunLoopMode beforeDate:[NSDate dateWithTimeIntervalSinceNow:1.0]];
    }
}

static void claimNowPlaying(void) {
    NSMutableDictionary *info = [NSMutableDictionary dictionary];
    info[MPNowPlayingInfoPropertyPlaybackRate] = @(0.0);
    info[MPNowPlayingInfoPropertyElapsedPlaybackTime] = @(0.0);
    [[MPNowPlayingInfoCenter defaultCenter] setNowPlayingInfo:info];
}

static void registerMediaKeyMonitor(void) {
    [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskSystemDefined handler:^(NSEvent *event) {
        if ([event type] == NSEventTypeSystemDefined && [event subtype] == 8) {
            int keyCode = (([event data1] & 0xFFFF0000) >> 16);
            int keyFlags = ([event data1] & 0x0000FFFF);
            int keyDown = ((keyFlags & 0xFF00) >> 8) == 0xA;
            if (keyDown) {
                switch (keyCode) {
                    case NX_KEYTYPE_PLAY:     goOnTogglePlayPause(); break;
                    case NX_KEYTYPE_NEXT:     goOnNext();            break;
                    case NX_KEYTYPE_PREVIOUS: goOnPrevious();        break;
                    case NX_KEYTYPE_FAST:     goOnNext();            break;
                    case NX_KEYTYPE_REWIND:   goOnPrevious();        break;
                }
            }
        }
    }];
}

// runCocoaMain must be called from the main OS thread.
// It initializes NSApplication, registers media key handlers, and runs the Cocoa event loop.
// It blocks until stopCocoaMain() is called.
static void runCocoaMain(void) {
    @autoreleasepool {
        NSApplication *app = [NSApplication sharedApplication];
        [app setActivationPolicy:NSApplicationActivationPolicyProhibited];
        setupRemoteCommandCenter();
        registerMediaKeyMonitor();
        claimNowPlaying();
        [app run];
    }
}

static void stopCocoaMain(void) {
    [[NSApplication sharedApplication] terminate:nil];
}

static void startRunLoop(void) {
    // legacy stub — not used when runCocoaMain() is called from main()
}

static void updateNowPlayingInfo(
    const char *title,
    const char *artist,
    const char *album,
    double durationSeconds,
    double elapsedSeconds,
    int isPlaying
) {
    NSMutableDictionary *info = [NSMutableDictionary dictionary];

    if (title) {
        info[MPMediaItemPropertyTitle] = [NSString stringWithUTF8String:title];
    }
    if (artist) {
        info[MPMediaItemPropertyArtist] = [NSString stringWithUTF8String:artist];
    }
    if (album) {
        info[MPMediaItemPropertyAlbumTitle] = [NSString stringWithUTF8String:album];
    }

    info[MPMediaItemPropertyPlaybackDuration] = @(durationSeconds);
    info[MPNowPlayingInfoPropertyElapsedPlaybackTime] = @(elapsedSeconds);
    info[MPNowPlayingInfoPropertyPlaybackRate] = isPlaying ? @(1.0) : @(0.0);

    [[MPNowPlayingInfoCenter defaultCenter] setNowPlayingInfo:info];
}

static void clearNowPlayingInfo(void) {
    [[MPNowPlayingInfoCenter defaultCenter] setNowPlayingInfo:nil];
}
*/
import "C"
import (
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/dece2183/yamusic-tui/media/handler"
)

var globalHandler *MacosHandler
var handlerMu sync.Mutex

type MacosHandler struct {
	msgMux  sync.Mutex
	msgChan chan handler.Message
	ansChan chan any
}

func init() {
	// Lock the main goroutine to the main OS thread so that
	// the Cocoa run loop (which requires the main thread) can be started here.
	runtime.LockOSThread()
}

func NewHandler(name, description string) *MacosHandler {
	mh := &MacosHandler{
		msgChan: make(chan handler.Message, 8),
		ansChan: make(chan any, 1),
	}

	handlerMu.Lock()
	globalHandler = mh
	handlerMu.Unlock()

	return mh
}

func (mh *MacosHandler) Start(handler func() error) error {
	var err error
	go func() {
		err = handler()
		C.teardownRemoteCommandCenter()
		C.clearNowPlayingInfo()
		C.stopCocoaMain()
	}()

	// Block the main OS thread running the Cocoa event loop.
	// This is required for MPRemoteCommandCenter and NSEvent monitoring to work.
	C.runCocoaMain()

	close(mh.msgChan)
	close(mh.ansChan)

	handlerMu.Lock()
	globalHandler = nil
	handlerMu.Unlock()

	return err
}

func (mh *MacosHandler) Message() <-chan handler.Message {
	return mh.msgChan
}

func (mh *MacosHandler) SendAnswer(ans any) {
	select {
	case mh.ansChan <- ans:
	default:
	}
}

func (mh *MacosHandler) OnEnded() {
	C.clearNowPlayingInfo()
}

func (mh *MacosHandler) OnVolume() {
}

func (mh *MacosHandler) OnPlayback() {
	mh.msgMux.Lock()
	mh.msgChan <- handler.Message{
		Type: handler.MSG_GET_METADATA,
	}
	metadata, ok := (<-mh.ansChan).(handler.TrackMetadata)
	mh.msgMux.Unlock()

	if !ok {
		return
	}

	title := metadata.Title
	artist := strings.Join(metadata.Artists, ", ")
	album := metadata.AlbumName

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	cArtist := C.CString(artist)
	defer C.free(unsafe.Pointer(cArtist))
	cAlbum := C.CString(album)
	defer C.free(unsafe.Pointer(cAlbum))

	C.updateNowPlayingInfo(
		cTitle, cArtist, cAlbum,
		C.double(metadata.Length.Seconds()),
		C.double(0),
		C.int(1),
	)
}

func (mh *MacosHandler) OnPlayPause() {
}

func (mh *MacosHandler) OnSeek(position time.Duration) {
}

func sendMsg(msg handler.Message) {
	handlerMu.Lock()
	mh := globalHandler
	handlerMu.Unlock()

	if mh == nil {
		return
	}

	mh.msgMux.Lock()
	select {
	case mh.msgChan <- msg:
	default:
	}
	mh.msgMux.Unlock()
}

//export goOnPlay
func goOnPlay() {
	sendMsg(handler.Message{Type: handler.MSG_PLAY})
}

//export goOnPause
func goOnPause() {
	sendMsg(handler.Message{Type: handler.MSG_PAUSE})
}

//export goOnTogglePlayPause
func goOnTogglePlayPause() {
	sendMsg(handler.Message{Type: handler.MSG_PLAYPAUSE})
}

//export goOnNext
func goOnNext() {
	sendMsg(handler.Message{Type: handler.MSG_NEXT})
}

//export goOnPrevious
func goOnPrevious() {
	sendMsg(handler.Message{Type: handler.MSG_PREVIOUS})
}

//export goOnSeek
func goOnSeek(positionSeconds C.double) {
	d := time.Duration(float64(positionSeconds) * float64(time.Second))
	sendMsg(handler.Message{Type: handler.MSG_SETPOS, Arg: d})
}

//export goOnSetVolume
func goOnSetVolume(volume C.double) {
	sendMsg(handler.Message{Type: handler.MSG_SET_VOLUME, Arg: float64(volume)})
}

//export goOnStop
func goOnStop() {
	sendMsg(handler.Message{Type: handler.MSG_STOP})
}
