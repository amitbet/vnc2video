# A **real wold** implementation of vnc client for go
After searching the web for an vnc client in golang which is not a toy & support more than handshake + RAW encoding, I came up blank, so, I set out to write one myself.

The video encoding part means that something can be viewed, and since I don't really feel like writing GTK UIs in 2018 (plus VNC viewers are a dime a dozen), a video file will do.
In actuality the images produced are go image structs and can easily be saved as JPEG, or displayed in any UI you want to create.

## Encoding support:
* Tight VNC
* Hextile
* ZLIB
* CopyRect
* Raw
* RRE
* ZRLE
* Rich-cursor pseudo
* Desktop Size Pseudo
* Cursor pos Pseudo

Since go has no good client UI library I chose to encode video instead, but the code is fully functional as a rfb renderer and **renders into golang native image.Image structs**.
## Video codec support:
* x264 (ffmpeg)
* dv8 (ffmpeg)
* dv9 (ffmpeg)
* MJpeg (native golnag)

## Frame Buffer Stream file support
* Supports reading & rendering fbs files that can be created by [vncProxy](https://github.com/amitbet/vncproxy)
* This allows recording vnc without the cost of video encoding while retaining the ability to have video later if the vnc session is marked as important.

## About
It may seem strange that I didn't use my previous vncproxy code in order to create this client, but since that code is highly optimized to be a proxy (never hold a full message in buffer & introduce no lags), it is not best suited to be a client, so instead of spending the time reverting all the proxy-specific code, I just started from the most advanced go vnc-client code I found.

Most of what I added is the rfb-encoder & video encoding implementations, there are naturally some additional changes in order to get a global canvas (draw.Image) to render on by all encodings.

The code for the encodings was gathered by peeking at several RFB source codes in cpp & some in java, reading the excellent documentation in [rfbproto](https://github.com/rfbproto/rfbproto/blob/master/rfbproto.rst), and **a lot** of gritty bit-plucking, pixel jogging & code cajoling until everything fell into place on screen.

I did not include tightPng in the supported encoding list since I didn't find a server to test it with, so I can't vouch for the previous implementation, If you have such a server handy, please check and tell me if it works.