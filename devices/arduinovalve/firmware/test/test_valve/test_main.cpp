//
// Created by taylojon on 9/13/2022.
//

#include <Arduino.h>
#include <unity.h>

#include "valve.h"
#include "config.h"

void setUp()
{

}
void tearDown()
{

}

void test_solenoid_open()
{
    uint8_t port = 0;
    port = port | SOLENOID_MASK;
    TEST_ASSERT_EQUAL(B00000111, port);
}

void test_solenoid_read()
{
    uint8_t port = B10101101;
    TEST_ASSERT_EQUAL(B101, port & SOLENOID_MASK);

}

void test_solenoid_write()
{
    uint8_t port = B10101000;
    port = (port & ~SOLENOID_MASK) | (SOLENOID_D & SOLENOID_MASK);
    TEST_ASSERT_EQUAL(B10101011, port);
}

void setup()
{
    delay(2000);
    UNITY_BEGIN();
    RUN_TEST(test_solenoid_open);
    RUN_TEST(test_solenoid_read);
    RUN_TEST(test_solenoid_write);
    UNITY_END();
}

void loop() {}