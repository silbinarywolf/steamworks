package steamworks_test

import (
	"log"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/silbinarywolf/steamworks"
	"github.com/silbinarywolf/steamworks/internal/steamerrors"
)

// steamId is set to Half-Life 2's Steam ID
const steamId = 220

const achievementRunUnitTests = "COMPLETE_LEVEL"

func TestRestartAppIfNecessary(t *testing.T) {
	onMainThread(func() {
		if !steamworks.SteamSupported {
			t.Skip("Steam is not supported on this platform. Skipping.")
		}
		if steamworks.RestartAppIfNecessary(steamId) {
			log.Fatal("RestartAppIfNecessary will return false because steam_appid.txt is in the same directory (which enables testing mode / stops restart app from running)")
		}
		// WARNING(Jake): 2019-09-27
		// If you are running Steam when you run this test, Steam might
		// prompt you with a warning saying "trying to run Half-Life 2 with certain logging flags"
		// or something to that affect.
	})
}

func TestInit(t *testing.T) {
	onMainThread(func() {
		if !steamworks.SteamSupported {
			t.Skip("Steam is not supported on this platform. Skipping.")
		}
		if steamworks.Init() {
			log.Fatalln("Steam should not initialize properly as tests shouldn't be launched on a machine with Steam running. We test for the failure case so this works on CI machines.")
		}
	})
}

func TestIsSteamRunning(t *testing.T) {
	onMainThread(func() {
		if !steamworks.SteamSupported {
			t.Skip("Steam is not supported on this platform. Skipping.")
		}
		if steamworks.IsSteamRunning() {
			log.Fatalln("Steam should not be running on the test box. This test will fail on your local machine if you are running Steam, regardless of whether it initialized or not properly.")
		}
	})
}

func TestRunCallbacks(t *testing.T) {
	onMainThread(func() {
		if !steamworks.SteamSupported {
			t.Skip("Steam is not supported on this platform. Skipping.")
		}
		if err := steamworks.RunCallbacks(); err != nil {
			log.Fatalf("RunCallbacks error: %v\n", err)
		}
	})
}

func TestGetAchievement(t *testing.T) {
	onMainThread(func() {
		if !steamworks.SteamSupported {
			t.Skip("Steam is not supported on this platform. Skipping.")
		}
		defer func() {
			if r := recover().(error); r != nil {
				if !strings.Contains(r.Error(), steamerrors.ErrNotInitialized.Error()) {
					t.Fatal("Expected GetAchievement to fail because Init has not been called successfully")
				}
			}
		}()
		hasAchievement := false
		if hasErr := steamworks.GetAchievement(achievementRunUnitTests, &hasAchievement); hasErr {
			t.Fatal("GetAchievement failed to execute.")
		}
		if hasAchievement {
			t.Logf("Has achievement")
		} else {
			t.Logf("Does not have achievement")
		}
	})
}

func TestSetAchievement(t *testing.T) {
	onMainThread(func() {
		if !steamworks.SteamSupported {
			t.Skip("Steam is not supported on this platform. Skipping.")
		}
		defer func() {
			if r := recover().(error); r != nil {
				if !strings.Contains(r.Error(), steamerrors.ErrNotInitialized.Error()) {
					t.Fatal("Expected SetAchievement to fail because Init has not been called successfully")
				}
			}
		}()
		if steamworks.SetAchievement(achievementRunUnitTests) {
			t.Fatal("SetAchievement failed to execute.")
		}
	})
}

func TestClearAchievement(t *testing.T) {
	onMainThread(func() {
		if !steamworks.SteamSupported {
			t.Skip("Steam is not supported on this platform. Skipping.")
		}
		defer func() {
			if r := recover().(error); r != nil {
				if !strings.Contains(r.Error(), steamerrors.ErrNotInitialized.Error()) {
					t.Fatal("Expected ClearAchievement to fail because Init has not been called successfully")
				}
			}
		}()
		if steamworks.ClearAchievement(achievementRunUnitTests) {
			t.Fatal("ClearAchievement failed to execute.")
		}
	})
}

// -----------------------------------------
// Force tests to execute on the main thread
// -----------------------------------------

// NOTE(Jake): 2019-09-27
// I'm not sure how *necessary* it is for all these functions
// to run on the mainthread, however since that's what I do in
// my project, I'm going to enforce it.

var mainfunc = make(chan func())

// onMainThread will execute the given function on the main thread
func onMainThread(f func()) {
	done := make(chan struct{})
	mainfunc <- func() {
		f()
		close(done)
	}
	<-done
}

func TestMain(m *testing.M) {
	done := make(chan int, 1)
	go func() {
		done <- m.Run()
	}()
	for {
		runtime.Gosched()
		select {
		case f := <-mainfunc:
			f()
		case res := <-done:
			os.Exit(res)
		default:
			// don't block if no message
		}
	}
}
