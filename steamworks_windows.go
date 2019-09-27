// +build windows

package steamworks

import (
	"reflect"
	"syscall"
	"unsafe"

	"github.com/silbinarywolf/steamworks/internal/steamerrors"
)

const (
	// SteamSupported will be true on platforms that support Steam
	SteamSupported = true
)

var (
	steamClientInterfaceVersion    = []byte("SteamClient017\x00")
	steamUserStatsInterfaceVersion = []byte("STEAMUSERSTATS_INTERFACE_VERSION011\x00")
)

var (
	steamUser      int32
	steamPipe      int32
	steamClient    uintptr
	steamUserStats uintptr
)

var (
	// Init
	hasInitialized bool

	// DLL
	steamApi syscall.Handle

	// GetProcAddress

	isSteamRunning                   uintptr
	runCallbacks                     uintptr
	steamInternal_CreateInterface    uintptr
	iSteamUserStats_SetAchievement   uintptr
	iSteamUserStats_GetAchievement   uintptr
	iSteamUserStats_ClearAchievement uintptr
)

const (
	uintptrSize uintptr = 0
	// NOTE(Jake): 2019-09-22
	// sizeOfSteamApiContext is incorrect, assuming an IntPtr is 8-bytes, this is probably more like 192 bytes
	// but Im just going to say its 512 so I have more than enough memory allocated for this
	// struct.
	// - unsafe.Sizeof(globals.callbackCounterAndContext)) == 8
	sizeOfSteamApiContext = 512
	steamApiContextSize   = 2 + (sizeOfSteamApiContext / unsafe.Sizeof(uintptrSize))
)

var (
	// NOTE(Jake): 2019-09-22
	// Allocate the memory for the SteamContext upfront so it *hopefully*
	// can't be garbage collected by Go.
	steamApiContextData       [steamApiContextSize]uintptr
	callbackCounterAndContext uintptr
	steamApiContext           *steamApiContextType
)

type steamApiContextType struct {
	m_pSteamClient             uintptr
	m_pSteamUser               uintptr
	m_pSteamFriends            uintptr
	m_pSteamUtils              uintptr
	m_pSteamMatchmaking        uintptr
	m_pSteamGameSearch         uintptr
	m_pSteamUserStats          uintptr
	m_pSteamApps               uintptr
	m_pSteamMatchmakingServers uintptr
	m_pSteamNetworking         uintptr
	m_pSteamRemoteStorage      uintptr
	m_pSteamScreenshots        uintptr
	m_pSteamHTTP               uintptr
	m_pController              uintptr
	m_pSteamUGC                uintptr
	m_pSteamAppList            uintptr
	m_pSteamMusic              uintptr
	m_pSteamMusicRemote        uintptr
	m_pSteamHTMLSurface        uintptr
	m_pSteamInventory          uintptr
	m_pSteamVideo              uintptr
	m_pSteamTV                 uintptr
	m_pSteamParentalSettings   uintptr
	m_pSteamInput              uintptr
}

func lazyDLLInit() {
	if steamApi == 0 {
		var err error
		steamApi, err = syscall.LoadLibrary("steam_api64.dll")
		if err != nil {
			panic(steamerrors.NewDLLError("lazyDLLInit", err))
		}
		isSteamRunning, _ = syscall.GetProcAddress(steamApi, "SteamAPI_IsSteamRunning")
		runCallbacks, _ = syscall.GetProcAddress(steamApi, "SteamAPI_RunCallbacks")
		steamInternal_CreateInterface, _ = syscall.GetProcAddress(steamApi, "SteamInternal_CreateInterface")
		iSteamUserStats_SetAchievement, _ = syscall.GetProcAddress(steamApi, "SteamAPI_ISteamUserStats_SetAchievement")
		iSteamUserStats_GetAchievement, _ = syscall.GetProcAddress(steamApi, "SteamAPI_ISteamUserStats_GetAchievement")
		iSteamUserStats_ClearAchievement, _ = syscall.GetProcAddress(steamApi, "SteamAPI_ISteamUserStats_ClearAchievement")
	}
}

func IsSteamRunning() bool {
	ret, _, _ := syscall.Syscall(uintptr(isSteamRunning), 0, 0, 0, 0)
	// NOTE(Jake): 2019-09-27
	// Ignoring error message as I get "The parameter is incorrect." for cases
	// where ret returns 0.
	// I've seen before even if I didn't get an actual error from the DLL-call so
	// I *think* it's safe to ignore.
	switch ret {
	case 0:
		return false
	case 1:
		return true
	default:
		panic(steamerrors.NewDLLBadReturnCodeError("IsSteamRunning", ret))
	}
	return true
}

func RunCallbacks() error {
	// NOTE(Jake): 2019-09-27
	// Ignoring "ret" value as this function is a "void" function
	_, _, callErr := syscall.Syscall(uintptr(runCallbacks), 0, 0, 0, 0)
	if callErr != 0 {
		return steamerrors.NewDLLError("RunCallbacks", callErr)
	}
	return nil
}

func RestartAppIfNecessary(appId uint32) bool {
	lazyDLLInit()
	restartAppIfNecessary, _ := syscall.GetProcAddress(steamApi, "SteamAPI_RestartAppIfNecessary")
	const nargs = 1
	ret, _, callErr := syscall.Syscall(
		uintptr(restartAppIfNecessary),
		nargs,
		uintptr(appId),
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("RestartAppIfNecessary", callErr))
	}
	switch ret {
	case 0:
		return false
	case 1:
		return true
	case 1163264:
		// NOTE(Jake): 2019-09-27
		// Seems like this returns a bad error code consistently.
		// I might be doing something wrong here... but it works on my machine
		// right now, so leaving this hack in.
		// Perhaps the correct logic here is just checking if "ret == 1"?
		return false
	default:
		panic(steamerrors.NewDLLBadReturnCodeError("RestartAppIfNecessary", ret))
	}
}

func Init() bool {
	if hasInitialized {
		return true
	}
	lazyDLLInit()
	steamApiInit, _ := syscall.GetProcAddress(steamApi, "SteamAPI_Init")
	const nargs = 0
	ret, _, callErr := syscall.Syscall(
		uintptr(steamApiInit),
		nargs,
		0,
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("Init", callErr))
	}
	if ret == 0 {
		return false
	}
	if ret != 1 {
		panic(steamerrors.NewDLLBadReturnCodeError("Init", ret))
	}
	// Initialize
	steamUser = getHSteamUser()
	steamPipe = getHSteamPipe()
	steamClient = createInterface(steamClientInterfaceVersion)
	steamUserStats = getISteamUserStats(steamClient, steamUser, steamPipe, steamUserStatsInterfaceVersion)

	// Setup SteamApiContext
	{
		newCallbackCounterAndContext := &steamApiContextData
		// NOTE(Jake): 2019-09-22
		// Documentation suggests this is a bad idea. Should NOT call Go functions from C as there will
		// be *problems*. However, im gonna do it anyway!
		newCallbackCounterAndContext[0] = uintptr(unsafe.Pointer(reflect.ValueOf(onContextInit).Pointer()))
		callbackCounterAndContext = uintptr(unsafe.Pointer(&newCallbackCounterAndContext[0]))
		steamApiContext = contextInit(callbackCounterAndContext)
	}
	steamApiContext.m_pSteamClient = steamClient
	steamApiContext.m_pSteamUserStats = steamUserStats

	hasInitialized = true
	return true
}

// GetAchievement will return true if the operation succeeded, otherwise it will return false
func GetAchievement(name string, achieved *bool) bool {
	if !hasInitialized {
		panic(steamerrors.ErrNotInitialized)
	}
	nameNullTerminated := []byte(name + "\x00")
	const nargs = 2
	ret, _, callErr := syscall.Syscall6(
		iSteamUserStats_GetAchievement,
		nargs,
		steamUserStats,
		uintptr(unsafe.Pointer(&nameNullTerminated[0])),
		uintptr(unsafe.Pointer(achieved)),
		0,
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("GetAchievement", callErr))
	}
	switch ret {
	case 0:
		return false
	case 1:
		return true
	default:
		// This can occur if "steamUserStats" is in an invalid state
		panic(steamerrors.NewDLLBadReturnCodeError("GetAchievement", ret))
	}
	return false
}

// SetAchievement will return true if the operation succeeded, otherwise it will return false
func SetAchievement(name string) bool {
	if !hasInitialized {
		panic(steamerrors.ErrNotInitialized)
	}
	nameNullTerminated := []byte(name + "\x00")
	const nargs = 2
	ret, _, callErr := syscall.Syscall6(
		iSteamUserStats_SetAchievement,
		nargs,
		steamUserStats,
		uintptr(unsafe.Pointer(&nameNullTerminated[0])),
		0,
		0,
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("SetAchievement", callErr))
	}
	switch ret {
	case 0:
		return false
	case 1:
		return true
	default:
		// This can occur if "steamUserStats" is invalid
		panic(steamerrors.NewDLLBadReturnCodeError("SetAchievement", ret))
	}
	return false
}

// ClearAchievement will return true if the operation succeeded, otherwise it will return false
func ClearAchievement(name string) bool {
	if !hasInitialized {
		panic(steamerrors.ErrNotInitialized)
	}
	nameNullTerminated := []byte(name + "\x00")
	const nargs = 2
	ret, _, callErr := syscall.Syscall6(
		iSteamUserStats_ClearAchievement,
		nargs,
		steamUserStats,
		uintptr(unsafe.Pointer(&nameNullTerminated[0])),
		0,
		0,
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("ClearAchievement", callErr))
	}
	switch ret {
	case 0:
		return false
	case 1:
		return true
	default:
		// This can occur if "steamUserStats" is invalid
		panic(steamerrors.NewDLLBadReturnCodeError("ClearAchievement", ret))
	}
	return false
}

func createInterface(name []byte) uintptr {
	const nargs = 1
	res, _, callErr := syscall.Syscall(
		steamInternal_CreateInterface,
		nargs,
		uintptr(unsafe.Pointer(&name[0])),
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("createInterface", callErr))
	}
	return res
}

// getHSteamUser number, starting at 1
func getHSteamUser() int32 {
	steamAPI_GetHSteamUser, _ := syscall.GetProcAddress(steamApi, "SteamAPI_GetHSteamUser")
	const nargs = 0
	ret, _, callErr := syscall.Syscall(
		steamAPI_GetHSteamUser,
		nargs,
		0,
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("getHSteamUser", callErr))
	}
	if ret == 0 {
		panic(steamerrors.NewDLLBadReturnCodeError("getHSteamUser", ret))
	}
	return int32(ret)
}

// getHSteamPipe number, starting at 1
func getHSteamPipe() int32 {
	steamAPI_GetHSteamPipe, _ := syscall.GetProcAddress(steamApi, "SteamAPI_GetHSteamPipe")
	const nargs = 0
	ret, _, callErr := syscall.Syscall6(
		steamAPI_GetHSteamPipe,
		nargs,
		0,
		0,
		0,
		0,
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("getHSteamPipe", callErr))
	}
	if ret == 0 {
		panic(steamerrors.NewDLLBadReturnCodeError("getHSteamPipe", ret))
	}
	return int32(ret)
}

func getISteamUserStats(instancePtr uintptr, steamUser int32, steamPipe int32, version []byte) uintptr {
	steamAPI_ISteamClient_GetISteamUserStats, _ := syscall.GetProcAddress(steamApi, "SteamAPI_ISteamClient_GetISteamUserStats")
	const nargs = 4
	ret, _, callErr := syscall.Syscall6(
		steamAPI_ISteamClient_GetISteamUserStats,
		nargs,
		instancePtr,
		uintptr(steamUser),
		uintptr(steamPipe),
		uintptr(unsafe.Pointer(&version[0])),
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("getISteamUserStats", callErr))
	}
	if ret == 0 {
		panic(steamerrors.NewDLLBadReturnCodeError("getISteamUserStats", ret))
	}
	return ret
}

func contextInit(callbackCounterAndContext uintptr) *steamApiContextType {
	steamInternal_ContextInit, _ := syscall.GetProcAddress(steamApi, "SteamInternal_ContextInit")
	const nargs = 1
	ret, _, callErr := syscall.Syscall(
		steamInternal_ContextInit,
		nargs,
		callbackCounterAndContext,
		0,
		0,
	)
	if callErr != 0 {
		panic(steamerrors.NewDLLError("contextInit", callErr))
	}
	if ret == 0 {
		panic(steamerrors.NewDLLBadReturnCodeError("contextInit", ret))
	}
	return (*steamApiContextType)(unsafe.Pointer(ret))
}

func onContextInit(contextPtr uintptr) {
	// NOTE(Jake): 2019-09-22
	// In my testing, it seems if you call Go functions from C
	// you end up crashing. However this WONT CRASH *if* I call
	// nothing from this function and just keep it empty.
	// I did try passing NULL / 0 instead of the address of a function
	// in "getCallbackCounterAndContext" but it failed
	//log.Printf("onContextInit: %v\n", contextPtr)

	// NOTE(Jake): 2019-09-22
	// This executes OK on Windows. Maybe just "log" has problems?
	// Or... maybe Go compiler just optimised this away?
	/*test := 1 + 1
	b := test + 1
	if b == 1 {

	}*/
}
