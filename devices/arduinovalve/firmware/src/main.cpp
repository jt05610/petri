//
// Created by taylojon on 6/22/2022.
//

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wunknown-pragmas"
#pragma ide diagnostic ignored "OCUnusedGlobalDeclarationInspection"

#include <Arduino.h>

#include "comm.h"
#include "valve.h"
#include "timing.h"
#include "config.h"
#include "protocol.h"

static struct
{
    uint8_t   buffer[BUFFER_CHARS];
    timing_t  params;
    message_t received;
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
    self.params.H = 0;
    comm_open();
    solenoid_open();
    timing_open(4000);
    self.received.buffer = self.buffer;
    self.received.size   = 0;
    Serial.print("ready\n");
}

void loop()
{
    comm_read(&self.received);
    if (self.received.size)
    {
        bool ok = process_message(&self.received, &self.params);
        format_response(ok, &self.received);
        if (ok)
        {
            timing_write(&self.params);
        }
        comm_write(&self.received);
    }
    Solenoid solenoid = timing_update();
    if (solenoid != nullptr)
    {
        solenoid_write(solenoid);
    }
}

#pragma clang diagnostic pop