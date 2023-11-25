NAME = server-status-go

build:
	cp -r webui/dist ./
	go build -v -trimpath -ldflags "-s -w" -o $(NAME) .
	upx -9 $(NAME)
