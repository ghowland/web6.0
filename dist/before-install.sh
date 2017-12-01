#!/bin/bash
#
# Script installed in the web6 package that is run prior to
# unpacking the package.

adduser --quiet --no-create-home --disabled-login --gecos "" web6 || true
