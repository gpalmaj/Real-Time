#!/bin/bash
cd costfunction/cost_fns/hall_request_assigner
echo "Building hall_request_assigner binary"
dmd main.d config.d elevator_algorithm.d elevator_state.d optimal_hall_requests.d d-json/jsonx.d -w -g -ofhall_request_assigner
cd ../../..

echo "Building go program"
go build -o heis .