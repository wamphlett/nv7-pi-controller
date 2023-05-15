# NV7 Pi Controller
A custom controller designed to work with the existing NV7 RGB controller hardware. Instead of controlling the RGB directly, button presses are sent via MQTT messages.

## Hardware requirements
This is not a plug and play solution, in order for the Raspberry Pi to be able to interface with the button PCB in the NV7, there is some additional circuitry that is required. The button PCB requires 5V to run and uses a resistor ladder to key the button inputs but the Raspberry Pi only has 3.3V GPIO pins and has no analong inputs. This means we need 2 solutions; a digitial switch which is able to switch the led line between ground and 5V and an [analog to digital converter](https://en.wikipedia.org/wiki/Analog-to-digital_converter) to read the button key. 

This is the schematic for the circuit I use to connect to the button PCB:

- 2x 10k ohm resisitors
- 1x 1k ohm resisitor
- 1x 5k ohm resisitor
- 1x 220 ohm resisitor
- 1x 2N3906 transistor
- 1x 2N3904 transistor
- 1x ADS1115 analog to digital converter

![schematic](docs/images/schematic.png "Schematic")

_note: I am by no means an electrical engineer, this circuit was put together through lots of trial and error. It works but there may be better/more efficient way to achieve the same thing. If by chance someone does have a better idea for this, I would love to hear your thoughts._