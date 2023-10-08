import unittest
from time import sleep

from autosampler import comm, config


class CommTestCase(unittest.TestCase):
    ser: comm.SerialInterface

    def setUp(self) -> None:
        self.ser = comm.SerialInterface(config.PORT, config.BAUD, config.TIMEOUT)
        self.assertEqual(
            b'G,10,Series 200 Autosampler,Version Rev 1.07,Serial No. 0',
            self.ser.initialize_port()
        )

    def tearDown(self) -> None:
        self.ser.close_port()

    def test_poll(self):
        print(self.ser.poll())

    def test_inject(self):
        params = comm.InjectionParameters(0, 10, 10, 250, 22, -10)
        for _ in range(0, 3):
            params._vial += 1
            self.ser.send_commands(params.commands)
            while self.ser.poll().unk_2 == 3:
                sleep(1)


if __name__ == '__main__':
    unittest.main()
