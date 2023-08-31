import RPi.GPIO as GPIO
from time import sleep

GPIO.setmode(GPIO.BOARD)

# pins = [31]
pins = [31, 32, 33, 35, 36, 37, 38, 40]


def cycle(pin, time):
    GPIO.output(pin, GPIO.HIGH)
    sleep(time)
    GPIO.output(pin, GPIO.LOW)
    sleep(time)


def main():
    for pin in pins:
        GPIO.setup(pin, GPIO.OUT, initial=GPIO.LOW)

    while True:
        for pin in pins:
            cycle(pin, .1)


if __name__ == "__main__":
    main()
