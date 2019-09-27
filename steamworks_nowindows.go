// +build !windows

package steamworks

import (
	"github.com/silbinarywolf/steamworks/internal/steamerrors"
)

const (
	// SteamSupported will be true on platforms that support Steam
	SteamSupported = false
)

func RestartAppIfNecessary(appId uint32) bool {
	panic(steamerrors.ErrNotSupported)
}

func Init() bool {
	panic(steamerrors.ErrNotSupported)
}

func IsSteamRunning() bool {
	panic(steamerrors.ErrNotSupported)
}

func RunCallbacks() error {
	panic(steamerrors.ErrNotSupported)
}

func GetAchievement(name string, achieved *bool) bool {
	panic(steamerrors.ErrNotSupported)
}

func SetAchievement(name string) bool {
	panic(steamerrors.ErrNotSupported)
}

func ClearAchievement(name string) bool {
	panic(steamerrors.ErrNotSupported)
}
