# Remote Controlled Radio Station (RCRS)

Made for our amateur radio club station to control and monitor the equipment remotely and see if everything works as intended.

## Configuration
All configuration options can be found in libs/myconst/myconst.go

## Running
Just run `go run .` in the folder where you found `README.md`.<br>
The first time it generates `pins.txt` based on `myconst.MAX_NUMBER_OF_PINS` then quits to let you configure the file.<br>
After that you can restart it and it'll start on port `8080` (if you haven't changed it)

If you want parallel port support you'll need to run with root privileges.

You can connect a camera anytime you want (before or even after starting the server), but don't disconnect it because the server will crash (can't do anything about it sadly).

There is an nginx config which is needed if you want to use the page outside of localhost. Listens on port `80`, proxies to `8080`<br>
It uses these paths:
- `/`: Main page (and only usable page)
- `/radio_ws`: websocket connection

## Technical information

### Abbreviations
#### Pin file:
- `T`: toggle button
- `P`: push button

### Syntax description
#### Pin file
```
<name of button>;<status: 0|1|->;<mode: T|P>
...
```

#### Websocket
First character denotes the command, remaining is the argument.
##### Client name
```
u*<name of client>
```
##### User list update
```
u[name of 1. user],...
```
##### Holding list update
```
h[<name of 1. user>;<button number>],...
```
##### Button status update
```
<status of 1. button><status of 2. button>...
```
##### Editor request (requesting client -> server)
```
e
```
##### Current editor
```
e[name of client]
```
##### JSON events
Command: `j`<br>
Args: JSON string with event wrapper (`{Event: "<event name>", Data: <event data>}`)
###### Page scheme data ("page_scheme")
```json
[                                       // each segment can contain multiple modules (Type = "buttons" | "cam")
	[
		{
			"Type": "buttons",          // button module
			"Buttons": [
				{
					"Name": "button 1", // name of button
					"Pin": 0,           // which pin to change on parallel port
					"Default": 0,       // default state (0: off, 1: on, -1: off+locked)
					"IsToggle": true    // 0: push button, 1: toggle button
				}
			]
		},
		{
			"Type": "cam",              // camera module
			"Name": "camera 1",         // name of camera
			"Device": "/dev/video0",    // camera's access path
			"Resolution": "1920x1080",  // camera's resolution
			"Fps": 30,                  // camera's fps
			"Format": "mjpeg"           // camera's video format ("mjpeg")
		}
	]
]
```