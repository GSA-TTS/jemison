#!/bin/bash

export SERVICE=extract

pushd /home/vcap/app/cmd/${SERVICE}
    echo Running the $SERVICE
    make run
popd