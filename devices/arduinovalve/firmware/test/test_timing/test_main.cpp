//
// Created by taylojon on 9/14/2022.
//

#include <Arduino.h>
#include <unity.h>

#include "timing.h"
#include "config.h"

void setUp()
{

}
void tearDown()
{

}

void test_timing_sum()
{
    timing_t params = {25, 25, 25, 25};
    uint8_t  sum    = params.A + params.B + params.C + params.D;
    TEST_ASSERT_EQUAL(100, sum);
}

void test_set_params()
{
    timing_t params = {25, 25, 25, 25};
    uint32_t period = 4000;
    uint32_t pulses[4];
    pulses[0] = params.A * (period / 100);
    TEST_ASSERT_EQUAL(1000, pulses[0]);
}

void setup()
{
    delay(2000);
    UNITY_BEGIN();
    RUN_TEST(test_timing_sum);
    RUN_TEST(test_set_params);
    UNITY_END();
}

void loop() {}