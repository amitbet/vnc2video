all: vnc2video.zip

ultra-ms-logon-2-encrypt: main.c d3des.h d3des.c
	gcc -Os -o $@ -m64 main.c
	strip $@

vnc2video: ../main.go ffmpeg
	(cd .. && go build -o $@)
	cp ../$@ .

vnc2video.zip: ultra-ms-logon-2-encrypt vnc2video ffmpeg
	zip -9 $@ $^
	sha256sum $@

ffmpeg:
	rm -rf ffmpeg-*-static*
	wget -q https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-64bit-static.tar.xz
	tar xf ffmpeg-release-64bit-static.tar.xz
	mv ffmpeg-*-static/ffmpeg .
	chmod +x ffmpeg
	rm -rf ffmpeg-*-static
	./ffmpeg -version

clean:
	rm -f vnc2video.zip vnc2video ultra-ms-logon-2-encrypt ffmpeg-*-static* ffmpeg *.mov

.PHONY: all clean
