#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "d3des.h"
#include "d3des.c"

static void vncEncryptBytes2(unsigned char *where, const int length, unsigned char *key) {
	int i, j;
	deskey(key, EN0);
	for (i = 0; i< 8; i++)
		where[i] ^= key[i];
	des(where, where);
	for (i = 8; i < length; i += 8) {
		for (j = 0; j < 8; j++)
			where[i + j] ^= where[i + j - 8];
		des(where + i, where + i);
	}
}

static void encrypt(unsigned char *key, int cipherTextSize, unsigned char *plainText) {
    unsigned char cipherText[cipherTextSize];
    memset(cipherText, 0, sizeof(cipherText));
    strcpy(cipherText, plainText);

    vncEncryptBytes2(cipherText, sizeof(cipherText), key);
    for (int i = 0; i < sizeof(cipherText); ++i) {
        printf("%02x", cipherText[i]);
    }
    printf("\n");
}

// usage:   ./ultra-ms-logon-2-encrypt des-key-hex cipher-text-size plain-text
// example: ./ultra-ms-logon-2-encrypt 3c89f4466dc2a67a 256 vagrant
int main(int argc, char **argv) {
    if (argc != 4) {
        return 1;
    }

    char *keyHex = argv[1];
    if (16 != strlen(keyHex)) {
        return 2;
    }
    unsigned char key[8];
    for (int i = 0; i < 8; ++i) {
        sscanf(keyHex, "%2hhx", &key[i]);
        keyHex += 2;
    }

    encrypt(key, atoi(argv[2]), argv[3]);

    return 0;
}
