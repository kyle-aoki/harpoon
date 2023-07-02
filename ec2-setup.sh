ssh $1 sudo apt update
ssh $1 sudo apt install docker.io -y
ssh $1 sudo chmod 666 /var/run/docker.sock
ssh $1 sudo apt install nginx -y
