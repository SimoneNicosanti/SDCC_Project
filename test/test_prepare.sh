## First upload all files

## Upload on S3 WITHOUT System
./system_start.sh 0

echo "UPLOADING TEST FILES"
docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 5mb.txt -e $2 -w NO
echo
docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 10mb.txt -e $2 -w NO
echo 
docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 20mb.txt -e $2 -w NO
echo
docker exec -t -i sdcc_project-client-1 python3 Main.py -u test -p test -o upload -f 30mb.txt -e $2 -w NO
echo

docker compose -f ../docker-compose.yml down