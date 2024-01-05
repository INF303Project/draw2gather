draw2gather: cmd/draw2gather/main.go
	go build -o draw2gather $<

deploy: draw2gather
	scp -i ./d2g-kp.pem draw2gather ubuntu@ec2-3-123-22-170.eu-central-1.compute.amazonaws.com:~/draw2gather

connect:
	ssh -i d2g-kp.pem ubuntu@ec2-3-123-22-170.eu-central-1.compute.amazonaws.com
