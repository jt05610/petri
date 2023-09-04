//
// Created by taylojon on 9/14/2022.
//

#include <Arduino.h>

#include "timing.h"
#include "valve.h"
#include "config.h"

static struct
{
    solenoid_t solenoid[8];
    uint32_t   period;
    uint32_t   pulses[8];
    timing_t *current_params;
    uint8_t    current_pulse;
    uint32_t   last_update;
    bool       running;
} self;

uint8_t masks[8] = {
        0x01 << VALVE_A,
        0x01 << VALVE_B,
        0x01 << VALVE_C,
        0x01 << VALVE_D,
        0x01 << VALVE_E,
        0x01 << VALVE_F,
        0x01 << VALVE_G,
        0x01 << VALVE_H

};

void
timing_open(uint32_t period)
{
    self.period        = period;
    for (uint32_t &pulse: self.pulses)
    {
        pulse = 0;
    }
    for (uint8_t  i    = 0; i < 8; i++)
    {
        if (i < 4)
        {
            self.solenoid[i].port = 0;
        } else
        {
            self.solenoid[i].port = 1;
        }
        self.solenoid[i].mask = masks[i];
    }
    self.current_pulse = 0;
    self.running       = false;
}

void
timing_close()
{
    timing_open(0);
}

uint32_t timing_sum(timing_t * params)
{
    return params->A + params->B + params->C + params->D +
           params->E + params->F + params->G + params->H;
}

void set_params(timing_t * params)
{
    uint32_t scale = self.period / timing_sum(params);
    self.current_params = params;
    self.pulses[0] = params->A * scale;
    self.pulses[1] = params->B * scale;
    self.pulses[2] = params->C * scale;
    self.pulses[3] = params->D * scale;
    self.pulses[4] = params->E * scale;
    self.pulses[5] = params->F * scale;
    self.pulses[6] = params->G * scale;
    self.pulses[7] = params->H * scale;
}

bool
timing_write(timing_t * params)
{
    uint8_t sum = timing_sum(params);
    if (sum == 0)
    {
        return false;
    }
    self.running = true;
    set_params(params);
    return true;
}

Solenoid
timing_update()
{
    Solenoid result  = nullptr;
    bool     updated = false;
    if (self.running)
    {
        if (millis() - self.last_update > self.pulses[self.current_pulse])
        {
            result = &self.solenoid[self.current_pulse];
            self.current_pulse = (self.current_pulse + 1) % 8;
            self.last_update   = millis();
            Serial.println(self.current_pulse);
        }
    } else
    {
        result = nullptr;
    }
    return result;
}

bool timing_set_period(uint32_t period)
{
    bool result = false;
    if (period > 0)
    {
        self.period = period;
        result      = true;
        timing_write(self.current_params);
    }
    return result;

}