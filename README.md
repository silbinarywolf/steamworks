# Steamworks API

[![Build Status](https://travis-ci.com/silbinarywolf/steamworks.svg?branch=master)](https://travis-ci.com/silbinarywolf/steamworks)
[![Documentation](https://godoc.org/github.com/silbinarywolf/steamworks?status.svg)](https://godoc.org/github.com/silbinarywolf/steamworks)
[![Report Card](https://goreportcard.com/badge/github.com/silbinarywolf/steamworks)](https://goreportcard.com/report/github.com/silbinarywolf/steamworks)

This is a Go implementation of the Steamworks API that aims to load the dynamic Steam API libraries without the use of CGo. 

This library was built so that I could utilize the synchronous functions such as `SetAchievement`, `ClearAchievement` and `GetAchievement`. You will will not be able to utilize methods that rely on callbacks as Go functions cannot be called from C-code without CGo.

If you're looking for a library with more API calls supported, consider looking at [BenLubar's Steamworks implementation](https://github.com/BenLubar/steamworks).

## Install

1) Install the library
```
go get github.com/silbinarywolf/steamworks
```

2) Copy DLL into your project directory (provided in the Steamworks SDK zip file)
```
steamworks_sdk_146.zip/sdk/redistributable_bin/win64/steam_api64.dll
```

3) Initialize Steamworks API
```go
const steamAppId = YOUR_STEAM_ID_HERE; // 220 = Half Life 2's Steam App ID

func main() {
	if steamworks.RestartAppIfNecessary(steamAppId) {
		log.Println("restarting with steam")
		os.Exit(1)
	}
	if !steamworks.Init() {
		log.Fatalln("Fatal Error - Steam must be running to play this game")
	}
}
```

4) Build your project directly into the Steam directory for testing
```
go build -o "D:\SteamLibrary\steamapps\common\\{YOUR_GAME}\game.exe" && /D/SteamLibrary/steamapps/common/{YOUR_GAME}/game
```


## Potential Problems

* Achievements only appear or trigger once the user quits the game. Not sure if this is normal behaviour or a quirk of initializing the Steam API without CGo.
* Clearing achievements will take effect immediately and not requiring quitting the game. If view your game on your Steam page, the achievement will become locked again immediately after you call `ClearAchievement`.
* `RestartAppIfNecessary` will return `1163264` in error cases rather than 0 or 1. Currently I'm unsure if this means I'm doing something wrong or if this value is expected.
* Currently tests are failing on Travis due to a "Class not registered" error that I cannot reproduce on my local machine. My research seems to suggest that it means a DLL it expects isn't [registered](https://web.archive.org/save/https://errorcodespro.com/heres-how-to-fix-the-class-not-registered-error/#What_Is_The_Meaning_Of_The_Class_Not_Registered_Error) however this specific DLL is not *meant* to be registered, so my hunch is that this error might be a symptom of Steam not being installed.


## Supported Operating Systems

* Windows

## Requirements

* Golang 1.13+

## Documentation

* [Documentation](https://godoc.org/github.com/silbinarywolf/steamworks)
* [License](LICENSE.md)

## Credits

* [FoxCouncil](https://github.com/FoxCouncil/Steamworks.Core) for their .NET implementation which I was able to use as a reference in the porting process.
