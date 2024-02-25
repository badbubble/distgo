#!/bin/bash

#SBATCH --job-name=go-build-timings   # Job name
#SBATCH --ntasks=1                        # Run a single task
#SBATCH --cpus-per-task=8                 # Number of CPU cores per task
#SBATCH --nodes=1                         # Request one node
#SBATCH --time=00:60:00                   # Time limit hrs:min:sec
#SBATCH --output=go_build_timings_%j.out  # Standard output and error log

# List of directory names
dir_names=("go-redis" "alist" "osmedeus")
time="/usr/bin/time"
go=/home/ppp23099/golang/go/bin/go

dir_names=("osmedeus") # ("go-redis" "alist" "osmedeus")
time="/usr/bin/time"
go=/home/ppp23099/golang/go/bin/go

# Loop over directory names
for dir in "${dir_names[@]}"; do
    echo "Entering directory $dir"
    cd "$dir"

    command_to_execute="$time $go build ."
    # We download the dependencies beforehand so it does not interfere with the build times.
    $go mod download

    # Consider no build cache
    # Loop over the values of GOMAXPROCS
    for gomaxprocs in 1 2 4 8; do
        echo "Running in $dir with GOMAXPROCS=$gomaxprocs, buildcache=disabled"

        # Iterate 10 times
        for i in {1..10}; do
          echo "Iteration: $i"
          $go clean -cache

          CGO_ENABLED=0 GOMAXPROCS=$gomaxprocs $command_to_execute
          echo ""
        done
    done

    # Consider build cache
    # Loop over the values of GOMAXPROCS
    for gomaxprocs in 1 2 4 8; do
        echo "Running in $dir with GOMAXPROCS=$gomaxprocs, buildcache=enabled"

        # Iterate 10 times
        for i in {1..10}; do
          echo "Iteration: $i"

          CGO_ENABLED=0 GOMAXPROCS=$gomaxprocs $command_to_execute
          echo ""
        done
    done

    # Go back to the original directory
    cd -
done