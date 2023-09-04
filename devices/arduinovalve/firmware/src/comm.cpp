//
// Created by taylojon on 9/13/2022.
//

#include <Arduino.h>
#include <HardwareSerial.h>

#include "../../fraction_collector/include/comm.h"
#include "../../fraction_collector/include/config.h"

static struct Comm
{
    uint8_t rx_buffer[BUFFER_CHARS];
    boolean new_data;
} comm;

void
comm_open()
{
    Serial.begin(BAUD);
    comm.new_data = false;
}

void
comm_read(message_t * message)
{
    static uint8_t len;
    uint8_t        read;
    message->size = 0;

    while(Serial.available() > 0 && !comm.new_data) {
        read = Serial.read();

        if (read != END_MARKER)
        {
            comm.rx_buffer[len++] = read;
            if (len >= BUFFER_CHARS)
                len = BUFFER_CHARS - 1;
        } else
        {
            comm.rx_buffer[len] = '\0';
            for (uint8_t i = 0; i < len; i ++)
                message->buffer[i] = comm.rx_buffer[i];
            message->size = len;
            len = 0;
            comm.new_data = true;
        }
    }
}

void comm_write(message_t * message)
{
    if (comm.new_data) {
        for (size_t i = 0; i < message->size; i ++)
            Serial.write(message->buffer[i]);
        Serial.write(END_MARKER);
        comm.new_data = false;
    }
}