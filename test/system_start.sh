#!/bin/bash
cd ../

docker compose down
docker compose up --scale edge=$1 --detach

## Balancer
echo
echo "Starting Balancer"
docker exec -t --workdir /src sdcc_project-load_balancer-1 /bin/bash /src/balancer_prp.sh
docker exec -td --workdir /src sdcc_project-load_balancer-1 ./balancer.out

## Registry
echo
echo "Starting Registry"
docker exec -t  sdcc_project-registry-1 /bin/bash /src/registry_prp.sh
docker exec -td  sdcc_project-registry-1 ./registry.out

## Edge
echo
echo "Starting Edges"

## Download packages for edges
for i in $( seq 1 $1 )
do
    docker exec -t  sdcc_project-edge-$i /bin/sh /src/edge_prp.sh
done

## Compile in docker volume
if [ $1 -ge 1 ]
then
    docker exec -t  sdcc_project-edge-1 go build -o edge.out main.go 
fi

## Start all edges
for i in $( seq 1 $1 )
do
    docker exec -td sdcc_project-edge-$i ./edge.out
done

sleep 5
echo
echo "###### System Started ######"
