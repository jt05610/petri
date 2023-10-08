//
// Created by taylojon on 9/13/2022.
//

#ifndef ECHO_VALVE_H
#define ECHO_VALVE_H

typedef struct solenoid_t * Solenoid;

typedef struct solenoid_t
{
    uint8_t id;
    uint8_t mask;
    uint8_t port;
}
                          solenoid_t;


void solenoid_open();

void solenoid_close();

Solenoid solenoid_read();

void solenoid_write(Solenoid solenoid);

#endif //ECHO_VALVE_H
