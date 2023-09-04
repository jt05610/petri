//
// Created by taylojon on 9/14/2022.
//

#ifndef FRACTION_COLLECTOR_TIMING_H
#define FRACTION_COLLECTOR_TIMING_H

#include "valve.h"

struct timing_t
{
    uint16_t A;
    uint16_t B;
    uint16_t C;
    uint16_t D;
    uint16_t E;
    uint16_t F;
    uint16_t G;
    uint16_t H;
};

void timing_open(uint32_t period);

bool timing_set_period(uint32_t period);

void timing_close();

timing_t timing_read();

bool timing_write(timing_t * params);

uint32_t timing_sum(timing_t * params);

Solenoid timing_update();

#endif //FRACTION_COLLECTOR_TIMING_H
