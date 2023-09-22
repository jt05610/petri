//
// Created by taylojon on 9/13/2022.
//
#include <Arduino.h>

#include "valve.h"
#include "config.h"

void
solenoid_open()
{
    DDRB = DDRB | PORT_AD_MASK;
    DDRD = DDRD | PORT_EH_MASK;
}

void
solenoid_close()
{
    DDRB = (DDRB & ~PORT_EH) | (0x00);
    DDRD = (DDRD & ~PORT_AD) | (0x00);
}

Solenoid
solenoid_read()
{
    return (Solenoid) (PORTB & SOLENOID_MASK);
}

void solenoid_write(Solenoid solenoid)
{
    PORTB = (PORTB & (~PORT_AD_MASK));
    PORTD = (PORTD & (~PORT_EH_MASK));
    if (solenoid->port == 1)
    {
        // clear the bits for the solenoid controlled by port B and set them according to the mask
        PORTB = PORTB | (solenoid->mask & PORT_AD_MASK);
    } else
    {
        // clear the bits for the solenoid controlled by port D and set them according to the mask
        PORTD = PORTD | (solenoid->mask & PORT_EH_MASK);
    }
}