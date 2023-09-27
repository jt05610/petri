//
// Created by taylojon on 9/14/2022.
//

#include "protocol.h"
#include "config.h"

static struct
{
    uint8_t tx_buffer[BUFFER_CHARS];
} self = {0};


// extract an uint16_t from the message buffer starting at the given index until we reach a comma, then convert the string to an integer
bool
extract_uint16(message_t * message, uint8_t * start, uint32_t * value)
{
    bool    result;
    uint8_t end = *start;
    while (end < message->size && message->buffer[end] != ',' &&
           message->buffer[end] != '\n')
    {
        end++;
    }
    if (end <= message->size)
    {
        char buffer[6];
        memcpy(buffer, message->buffer + *start, end - *start);
        buffer[end - *start] = '\0';
        *value = strtol(buffer, nullptr, 10);
        *start = end + 1;
        result = true;
    } else
    {
        result = false;
    }

    return result;
}

// the run message is a message with an R header then 16 bytes of data where each byte is a 2 digit hex number. The first 8 bytes are the A-D
// valves and the second 8 bytes are the E-H valves. The first byte of each pair is the low byte and the second byte is the high byte.
bool
process_run(message_t * message, timing_t * params)
{
    uint32_t * values[8] = {
            &params->A,
            &params->B,
            &params->C,
            &params->D,
            &params->E,
            &params->F,
            &params->G,
            &params->H
    };
    uint8_t current = 1;
    for (auto &value: values)
    {
        if (!extract_uint16(message, &current, value))
        {
            Serial.println(
                    "error: could not extract uint16_t starting from character " +
                    String(current));
            break;
        }
    }
    return true;
}

/*
 * Example: <S>
 */
bool process_stop(message_t * message, timing_t * params)
{
    bool result;
    if (message->size != 1)
    {
        result = false;
    } else
    {
        params->A = 0;
        params->B = 0;
        params->C = 0;
        params->D = 0;
        params->E = 0;
        params->F = 0;
        params->G = 0;
        params->H = 0;
        result = true;
    }
    return result;
}

bool
process_message(message_t * message, timing_t * params)
{
    bool result;
    switch (message->buffer[0])
    {
        case RUN_HEADER:
            result = process_run(message, params);
            break;
        case STOP_HEADER:
            result = process_stop(message, params);
            break;
        case PERIOD_HEADER:
            uint8_t  start;
            uint32_t period;
            start  = 1;
            result = extract_uint16(message, &start, &period);
            if (result)
            {
                timing_set_period(period);
                result = false;
            } else
            {
                Serial.println("error: could not extract period");
            }
            break;
        default:
            Serial.println(
                    "error: unknown header: " + String(message->buffer[0]));
            result = false;
            break;
    }
    return result;
}

void format_response(bool status, message_t * buffer)
{
    if (status)
    {
        buffer->buffer[0] = 'o';
        buffer->buffer[1] = 'k';
        buffer->size = 2;
    } else
    {

        buffer->buffer[0] = 'e';
        buffer->buffer[1] = 'r';
        buffer->buffer[2] = 'r';
        buffer->buffer[3] = 'o';
        buffer->buffer[4] = 'r';
        buffer->size = 5;
    }
}
