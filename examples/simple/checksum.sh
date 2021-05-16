#!/bin/sh

sha256sum some.dat | sha256sum - | cut -b 1-6
