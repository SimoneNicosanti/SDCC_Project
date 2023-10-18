#!/bin/bash

MAX_RUN=3
# Files For Cache
dd if=/dev/zero of=./client_files/10mb.txt bs=1M count=10
dd if=/dev/zero of=./client_files/20mb.txt bs=1M count=20
dd if=/dev/zero of=./client_files/30mb.txt bs=1M count=30

# Other file sizes
dd if=/dev/zero of=./client_files/100mb.txt bs=1M count=100
dd if=/dev/zero of=./client_files/250mb.txt bs=1M count=250
dd if=/dev/zero of=./client_files/500mb.txt bs=1M count=500
dd if=/dev/zero of=./client_files/750mb.txt bs=1M count=750
dd if=/dev/zero of=./client_files/1000mb.txt bs=1M count=1000

## -u username -p password -o operazione -f file_name
## Non so se vogliamo eseguire i test in parallelo: in caso aggiungere una & a fine comando e una sleep a fine cicli per non sovraccaricare il sistema

# Test Upload
if [ $1 = "up" ]
then
    echo "TEST UPLOAD"
    for i in $( seq 0 $MAX_RUN )
    do
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 10mb.txt
        echo 
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 20mb.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 30mb.txt
        echo

        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 100mb.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 250mb.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 500mb.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 750mb.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 1000mb.txt
        echo
    done
fi

# Test Download cachable files
if [ $1 = "down_cache" ]
then
    echo "TEST WITH CACHE FILES"
    ## First upload all files
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 10mb.txt
    echo 
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 20mb.txt
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 30mb.txt
    echo

    for i in $( seq 0 $MAX_RUN )
    do
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 10mb_test.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 20mb_test.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 30mb_test.txt
        echo
    done
fi

# Test Download no cachable files
if [ $1 = "down_no_cache" ]
then
    echo "TEST DOWNLOAD NO CACHE"
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 100mb.txt
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 250mb.txt
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 500mb.txt
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 750mb.txt
    echo
    docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 1000mb.txt
    echo

    for i in $( seq 0 $MAX_RUN )
    do
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 100mb_test.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 250mb_test.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 500mb_test.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 750mb_test.txt
        echo
        docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 1000mb_test.txt
        echo
    done
fi

# Test Delete --> Come la facciamo??
# for i in $( seq 0 $MAX_RUN )
# do
#     echo
#     echo RUN TEST $i
#     docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 100mb_test.txt

#     docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 100mb.txt
#     docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 250mb.txt
#     docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 500mb.txt
#     docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 750mb.txt
#     docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o download -f 1000mb.txt
    
# done

# To avoid commit problems
rm ./client_files.*mb.txt