build: 
	go build -o blockchain_go . 

commit: 
	git add . && git commit -m "New changes"

push: commit
	git push origin