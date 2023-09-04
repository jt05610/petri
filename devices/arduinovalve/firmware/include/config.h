//
// Created by taylojon on 9/13/2022.
//

#ifndef ECHO_CONFIG_H
#define ECHO_CONFIG_H

#define BUFFER_CHARS 100
#define BAUD 115200

#define END_MARKER '\n'

#define PORT_AD PORTD
#define PORT_EH PORTB
#define SOLENOID_MASK 0x07

#define VALVE_A 2
#define VALVE_B 3
#define VALVE_C 4
#define VALVE_D 5

#define VALVE_E 0
#define VALVE_F 1
#define VALVE_G 2
#define VALVE_H 3

#define PORT_AD_MASK ((0x01 << VALVE_A) | (0x01 << VALVE_B) | (0x01 << VALVE_C) | (0x01 << VALVE_D))
#define PORT_EH_MASK ((0x01 << VALVE_E) | (0x01 << VALVE_F) | (0x01 << VALVE_G) | (0x01 << VALVE_H))
#define PORT_MASK (PORT_AD_MASK | PORT_EH_MASK)

#endif //ECHO_CONFIG_H
