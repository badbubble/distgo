#!/bin/bash

#SBATCH --job-name=go-build-pprof   # Job name
#SBATCH --ntasks=1                        # Run a single task
#SBATCH --cpus-per-task=8                 # Number of CPU cores per task
#SBATCH --nodes=1                         # Request one node
#SBATCH --time=00:60:00                   # Time limit hrs:min:sec
#SBATCH --output=go_build_pprof_%j.out  # Standard output and error log

# List of directory names
dir_names=("go-redis" "alist" "osmedeus")
time="/usr/bin/time"
go=/home/ppp23099/golang/go/bin/go

# Loop over directory names
for dir in "${dir_names[@]}"; do
    echo "Entering directory $dir"
    cd "$dir"

    command_to_execute="$time $go build"
    # We download the dependencies beforehand so it does not interfere with the build times.
    $go mod download
    # (Optionally) We execute the build command so that we do not skew the results when using a cache.
    # CGO_ENABLED=1 GOMAXPROCS=4 $command_to_execute $location

    # Consider build cache
    # Loop over the values of GOMAXPROCS
    for gomaxprocs in 1 2 4 8; do
        echo "Running in $dir with GOMAXPROCS=$gomaxprocs"

        # (Optionally) We clear the cache after every build if we want to start over every iteration.
        $go clean -cache

        trace="-gcflags=-cpuprofile=/home/ppp23099/golang/results/pprof-$dir-$gomaxprocs-no-cache"
        output="-o /home/ppp23099/golang/results/build-$dir-$gomaxprocs"
        location="."

        CGO_ENABLED=0 GOMAXPROCS=$gomaxprocs $command_to_execute $output $trace $graph $location

        rm home/ppp23099/golang/results/pprof-$dir-$gomaxprocs
    done

    # Go back to the original directory
    cd -
done