# Solargo

[![Build Status](https://github.com/teuron/SolarGo/workflows/run%20tests/badge.svg)](https://github.com/teuron/SolarGo/actions?query=workflow%3A%22run+tests%22)
[![Coverage Status](https://coveralls.io/repos/github/teuron/solargo/badge.svg?branch=master)](https://coveralls.io/github/teuron/solargo?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/teuron/SolarGo)](https://goreportcard.com/report/github.com/teuron/SolarGo)

Solargo is a solar display for everyone who has a home solarplant with a Fronius Symo inverter. For now, we only support the Fronius Symo inverter series, but feel free to open a PR to add more.


Usage
----

Clone the repository:

    git clone https://github.com/teuron/solargo.git

Change to the folder:

    cd solargo

Install golang and make:

    sudo apt-get install golang make

Edit solargo.service (adapt WorkingDirectory, ExecStart and User) and config_empty.yaml

Copy config_empty.yaml to config.yaml:

    cp config_empty.yaml config.yaml

Run solargo as a systemd service:

    make all

Enjoy!


License
----

This project was created under the [MIT license][8]


[8]: LICENSE
