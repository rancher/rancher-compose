#!/bin/bash
gsutil -m cp -p winged-math-749  -r dist/artifacts/v*  gs://releases.rancher.com/compose/beta
