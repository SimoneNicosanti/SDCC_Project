#!/bin/bash

./system_start.sh $2

cd ../docker
MAX_RUN=5
CACHABLE_REQ_NUM=3
NO_CACHABLE_REQ_NUM=2
# Files For Cache
dd if=/dev/zero of=./client_files/5mb.txt bs=1M count=5
dd if=/dev/zero of=./client_files/10mb.txt bs=1M count=10
dd if=/dev/zero of=./client_files/20mb.txt bs=1M count=20
dd if=/dev/zero of=./client_files/30mb.txt bs=1M count=30

# Other file sizes
dd if=/dev/zero of=./client_files/100mb.txt bs=1M count=100
dd if=/dev/zero of=./client_files/250mb.txt bs=1M count=250

## Flag meaning
# -u username 
# -p password 
# -o operazione 
# -f file_name
# -e edge_num
# -w time_output_file

# Test Upload
if [ $1 = "up" ]
then
    echo "TEST UPLOAD"
    for i in $( seq 1 $MAX_RUN )
    do
        # CACHABLE
        for j in $(seq 1 $CACHABLE_REQ_NUM)
        do
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 5mb.txt -e $2  -w Parallel.csv &
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 10mb.txt -e $2 -w Parallel.csv & 
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 20mb.txt -e $2 -w Parallel.csv &
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 30mb.txt -e $2 -w Parallel.csv &
        done

        for j in $( seq 1 $NO_CACHABLE_REQ_NUM)
        do
        # NO CACHABLE
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 100mb.txt -e $2 -w Parallel.csv &
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 250mb.txt -e $2 -w Parallel.csv &
        done

        ## To wait termination of previous requests
        sleep 15
    done
fi


## Test Download
if [ $1 = "down_all" ]
then

    echo "UPLOADING ALL FILES"
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 5mb.txt -e $2 -w NO
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 10mb.txt -e $2 -w NO
    echo 
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 20mb.txt -e $2 -w NO
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 30mb.txt -e $2 -w NO
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 100mb.txt -e $2 -w NO
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 250mb.txt -e $2 -w NO
    echo

    ## Remove from edge caches
    for i in $( seq 1 $2 )
    do
        docker exec -t  sdcc_project-edge-$i /bin/sh rm /files/*
    done

    echo "TEST DOWNLOAD"
    for i in $( seq 1 $MAX_RUN )
    do
        for j in $(seq 1 $CACHABLE_REQ_NUM)
        do
            # CACHABLE
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 5mb_test.txt -e $2  -w Parallel.csv &
            echo
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 10mb_test.txt -e $2 -w Parallel.csv &
            echo 
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 20mb_test.txt -e $2 -w Parallel.csv &
            echo
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 30mb_test.txt -e $2 -w Parallel.csv &
            echo
        done

        # NO CACHABLE
        for j in $( seq 1 $NO_CACHABLE_REQ_NUM)
        do
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 100mb_test.txt -e $2 -w Parallel.csv &
            echo
            docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 250mb_test.txt -e $2 -w Parallel.csv &
            echo
        done

        sleep 15
    done
fi



# To avoid commit problems
rm ./client_files/*