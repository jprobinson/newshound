
deploy:
	@docker build --tag gcr.io/newshound/np .; \
	gcloud docker  -- push gcr.io/newshound/np:latest;
	gcloud app deploy --project newshound --image-url gcr.io/newshound/np:latest app.yaml
