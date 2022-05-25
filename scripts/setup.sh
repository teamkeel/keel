#!/bin/bash

# Setup git precommit hook
echo 'Setting up git precommit hook...'
brew install pre-commit -q
pre-commit install --hook-type commit-msg

echo 'Finished setup!'
