build: 
	go build -o blockchain_go . 

pull: 
	git pull 

commit: pull
	git add . && git commit -m "New changes"

push: commit
	git push origin