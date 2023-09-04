//
// Created by taylojon on 9/13/2022.
//

#ifndef ECHO_COMM_H
#define ECHO_COMM_H

#include "protocol.h"

void comm_open();

void comm_read(message_t * buffer);

void comm_write(message_t * message);

#endif //ECHO_COMM_H
