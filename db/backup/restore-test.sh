#!/bin/bash

gunzip < $1 | psql --echo-errors -h localhost -U postgres -X project_manager_test
