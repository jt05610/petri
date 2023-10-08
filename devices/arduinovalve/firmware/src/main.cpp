//
// Created by taylojon on 6/22/2022.
//

#include <Arduino.h>

#include "comm.h"
#include "valve.h"
#include "timing.h"
#include "config.h"
#include "protocol.h"

static struct
{
    uint8_t   buffer[BUFFER_CHARS];
    uint8_t   tx_buffer[BUFFER_CHARS];
    timing_t  params;
    message_t received;
    message_t response;
} self;

void setup()
{
    self.params.A = 0;
    self.params.B = 0;
    self.params.C = 0;
    self.params.D = 0;
    self.params.E = 0;
    self.params.F = 0;
    self.params.G = 0;
    comm_open();
    solenoid_open();
    timing_open(4000);
    self.received.buffer = self.buffer;
    self.response.buffer = self.tx_buffer;
    self.response.size   = 0;
    self.received.size   = 0;
    Serial.print("ready\n");
}

__attribute__((unused)) void loop()
{
    comm_read(&self.received);
    if (self.received.size)
    {
        bool ok = process_message(&self.received, &self.params);
        format_response(ok, &self.response);
        if (ok)
        {
            timing_write(&self.params);
        }
        comm_write(&self.response);
    }
    Solenoid solenoid = timing_update();
    if (solenoid != nullptr)
    {
        solenoid_write(solenoid);
    }
}
