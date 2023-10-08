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
    uint32_t   pulses[N_PULSES];
    uint32_t   active_pulses[N_PULSES];
    solenoid_t active_solenoids[N_PULSES];
    uint8_t    n_pulses;
    timing_t * current_params;
    uint8_t  current_pulse;
    uint32_t last_update;
    bool     running;
} self;

uint8_t masks[7] = {
        0x01 << VALVE_A,
        0x01 << VALVE_B,
        0x01 << VALVE_C,
        (0x01 << VALVE_E | 0x01 << VALVE_D),
        (0x01 << VALVE_F | 0x01 << VALVE_D),
        (0x01 << VALVE_G | 0x01 << VALVE_D),
        (0x01 << VALVE_H | 0x01 << VALVE_D),
};

void
timing_open(uint32_t period)
{
    self.current_params = nullptr;
    self.period         = period;
    for (uint32_t &pulse: self.pulses)
    {
        pulse = 0;
    }
    for (uint32_t &pulse: self.active_pulses)
    {
        pulse = 0;
    }

    self.n_pulses = 0;
    for (uint8_t i     = 0; i < N_PULSES; i++)
    {
        if (i < 3)
        {
            self.solenoid[i].port = 1;
        } else
        {
            self.solenoid[i].port = 0;
        }
        self.solenoid[i].mask = masks[i];
        self.solenoid[i].id   = i + 'A';
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
           params->E + params->F + params->G;
}

void set_params(timing_t * params)
{
    self.n_pulses  = 0;
    uint32_t total = timing_sum(params);
    self.current_params = params;
    self.pulses[0] = self.period * (1000 * params->A / total);
    self.pulses[1] = self.period * (1000 * params->B / total);
    self.pulses[2] = self.period * (1000 * params->C / total);
    self.pulses[3] = self.period * (1000 * params->D / total);
    self.pulses[4] = self.period * (1000 * params->E / total);
    self.pulses[5] = self.period * (1000 * params->F / total);
    self.pulses[6] = self.period * (1000 * params->G / total);

    for (uint8_t i = 0; i < N_PULSES; i++)
    {
        if (self.pulses[i] > 0)
        {
            self.active_pulses[self.n_pulses]    = self.pulses[i];
            self.active_solenoids[self.n_pulses] = self.solenoid[i];
            self.n_pulses++;
        }
    }
}

bool
timing_write(timing_t * params)
{
    uint64_t sum = timing_sum(params);
    if (sum == 0)
    {
        self.running = false;
        return false;
    }
    self.running = true;
    set_params(params);
    return true;
}

Solenoid
timing_update()
{
    Solenoid result     = nullptr;
    uint8_t  last_pulse = self.current_pulse;
    uint32_t now;
    if (self.running)
    {
        now = micros();
        if (now - self.last_update > self.active_pulses[self.current_pulse])
        {
            result = &self.active_solenoids[self.current_pulse];
            self.current_pulse = (self.current_pulse + 1) % self.n_pulses;
            self.last_update   = now;
            if (self.current_pulse != last_pulse)
            {
                Serial.print("opened: ");
                Serial.println(
                        (char) self.active_solenoids[self.current_pulse].id);
            }
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
        result = true;
        if (self.current_params != nullptr)
        {
            timing_write(self.current_params);
        }
    }
    return result;
}
