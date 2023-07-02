ssh $1 "rm -rf harpoon"
GOOS=linux GOARCH=amd64 go build
scp harpoon $1:~/harpoon
rm harpoon
