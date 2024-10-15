#include <stdio.h>
#include <unistd.h>
#include <sys/io.h>
#include <math.h>

#define base 0x378 // Base address of the parallel port

void enable_perm() {
    if (ioperm(base, 1, 1)) {
        printf("Access denied to %x\n", base);
    }
}

void disable_perm() {
    if (ioperm(base, 1, 0)) {
        printf("Couldn't revoke access to %x\n", base);
    }
}

void set_pin(int pin, int data, int voltage) {
    if (voltage) {
        data ^= (int)pow(2, pin);
    } else {
        data &= ~((int)pow(2, pin));
    }

    printf("%x\n", data);

    // outb(data, base);
}
