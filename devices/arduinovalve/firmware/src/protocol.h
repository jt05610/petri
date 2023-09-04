//
// Created by taylojon on 9/14/2022.
//

#ifndef FRACTION_COLLECTOR_PROTOCOL_H
#define FRACTION_COLLECTOR_PROTOCOL_H

#include <Arduino.h>
#include "timing.h"

#define RUN_HEADER 'R'
#define STOP_HEADER 'S'
#define PERIOD_HEADER 'P'

struct message_t
{
    uint8_t * buffer;
    size_t size;
};

bool process_message(message_t *message, timing_t * params);

void format_response(bool status, message_t * buffer);

#endif //FRACTION_COLLECTOR_PROTOCOL_H
