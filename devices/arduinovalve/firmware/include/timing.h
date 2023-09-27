//
// Created by taylojon on 9/14/2022.
//

#ifndef FRACTION_COLLECTOR_TIMING_H
#define FRACTION_COLLECTOR_TIMING_H

#include "valve.h"

struct timing_t
{
    uint32_t A;
    uint32_t B;
    uint32_t C;
    uint32_t D;
    uint32_t E;
    uint32_t F;
    uint32_t G;
    uint32_t H;
};

void timing_open(uint32_t period);

bool timing_set_period(uint32_t period);

void timing_close();

timing_t timing_read();

bool timing_write(timing_t * params);

uint32_t timing_sum(timing_t * params);

Solenoid timing_update();

#endif //FRACTION_COLLECTOR_TIMING_H
