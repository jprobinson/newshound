build:
	@docker run -it --rm -v "$$(pwd):/data" -w /data digitallyseamless/nodejs-bower-grunt npm install
	@docker run -it --rm -v "$$(pwd):/data" -w /data digitallyseamless/nodejs-bower-grunt bower install 

run:
	@docker run -it --rm -p 35729:35729 -p 9000:9000 -v "$$(pwd):/data" -w /data digitallyseamless/nodejs-bower-grunt grunt serve
