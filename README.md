# A **real wold** implementation of vnc client for go
After searching the web for an vnc client in golang which is not a toy & support more than handshake + RAW encoding, I came up blank, so, I set out to write it myself.
## Encoding support:
* Tight VNC
* Hextile
* ZLIB
* CopyRect
* Raw
* RRE
* ZRLE [**TBD - coming soon**]
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
It may seem strange that i didn't use my previous vncproxy code in order to create this client, but since that code is highly optimized to be a proxy (never hold a full message in buffer & introduce no lags), it is not best suited to be a client, so instead of spending the time reverting all the proxy-specific code, I just started from the most advanced go vnc-client code I found.

Most of what I added is the rfb-encoder & video encoding implementations, there are naturally some additional changes in order to get a global canvas (draw.Image) to render on by all encodings.

The code for the encodings was gathered by looking at several RFB source codes in cpp & some in java, reading the excellent documentation in [rfbproto](https://github.com/rfbproto/rfbproto/blob/master/rfbproto.rst), and **a lot** of gritty bit-plucking, pixel jogging & code cajoling until everything fell into place on screen.