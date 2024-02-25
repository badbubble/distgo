#!/bin/bash

#SBATCH --job-name=go-build-actiongraph   # Job name
#SBATCH --ntasks=1                        # Run a single task
#SBATCH --cpus-per-task=8                 # Number of CPU cores per task
#SBATCH --nodes=1                         # Request one node
#SBATCH --time=00:60:00                   # Time limit hrs:min:sec
#SBATCH --output=go_build_actiongraph_%j.out  # Standard output and error log

# List of directory names
dir_names=("go-redis" "alist" "osmedeus")
time="/usr/bin/time"
go=/home/ppp23099/golang/go/bin/go

# Loop over directory names
for dir in "${dir_names[@]}"; do
    echo "Entering directory $dir"
    cd "$dir"

    # Determine the command to execute based on the directory name or other criteria
    # This is a placeholder command assignment; customize it as needed
    if [ -f "special_program.go" ]; then
        command_to_execute="go run special_program.go"
    else
        command_to_execute="$time $go build"
    fi

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

        graph="-debug-actiongraph=/home/ppp23099/golang/results/actiongraph-$dir-$gomaxprocs-no-cache"
        trace="-debug-trace=/home/ppp23099/golang/results/trace-$dir-$gomaxprocs-no-cache"
        output="-o /home/ppp23099/golang/results/build-$dir-$gomaxprocs"
        location="."

        CGO_ENABLED=0 GOMAXPROCS=$gomaxprocs $command_to_execute $trace $graph $location

        rm home/ppp23099/golang/results/build-$dir-$gomaxprocs
    done

    # Go back to the original directory
    cd -
done