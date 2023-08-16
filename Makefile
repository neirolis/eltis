# ELTIS БЛОК ВЫЗОВА DP5000.В2-хх

NAME=eltis

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ${NAME}


zip: build
	zip ${NAME}.zip ${NAME} ${NAME}.service

deploy: zip
	rsync -rvtzz ${NAME}.zip rtmip.info:/srv/packages/devices/