# About

Encrypts a username and password as the ultra-vnc ms-auto-logon-2 plugin.

# Usage

Build:

```bash
make clean all
```

Then use it like:

```bash
./vnc2video example1.windows.com:5900 vagrant vagrant
```

## Source Code

Most of the included source code came from:

* https://github.com/TigerVNC/tigervnc/blob/8c6c584377feba0e3b99eecb3ef33b28cee318cb/common/rfb/d3des.h
* https://github.com/TigerVNC/tigervnc/blob/8c6c584377feba0e3b99eecb3ef33b28cee318cb/common/rfb/d3des.c
