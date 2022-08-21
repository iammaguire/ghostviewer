# ghostviewer
Covert remote desktop POC for Windows. Uses DXGI's DDAPI for capture on Windows 8 and above. This isn't available on ealier versions of Windows so BitBlt is used in these cases. Keyboard and mouse input supported, except for dragging.

Note that keyboard and mouse IO is commented out by default. To enable it check io/iodriver.go. This is because when the client and server is run on the same machine the mouse is glitched around by the "loopback" messaging, making the computer impossible to use until the application has exited.
