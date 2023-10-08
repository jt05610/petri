//
// Created by taylojon on 9/13/2022.
//

#ifndef ECHO_CONFIG_H
#define ECHO_CONFIG_H

#define BUFFER_CHARS 100
#define BAUD 115200

#define END_MARKER '\n'

#define PORT_AD PORTB
#define PORT_EH PORTD
#define SOLENOID_MASK 0x07

#define VALVE_A 3
#define VALVE_B 2
#define VALVE_C 1

#define VALVE_D 6
#define VALVE_E 5
#define VALVE_F 4
#define VALVE_G 3
#define VALVE_H 2

#define PORT_AD_MASK ((0x01 << VALVE_A) | (0x01 << VALVE_B) | (0x01 << VALVE_C))
#define PORT_EH_MASK ((0x01 << VALVE_D) | (0x01 << VALVE_E) | (0x01 << VALVE_F) | (0x01 << VALVE_G) | (0x01 << VALVE_H))
#define PORT_MASK (PORT_AD_MASK | PORT_EH_MASK)

#endif //ECHO_CONFIG_H
