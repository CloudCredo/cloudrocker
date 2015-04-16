TAG?=latest

all: rocker-ci

push:
		docker push hatofmonkeys/rocker-ci

rocker-ci: rocker-ci/Dockerfile
		docker build -t hatofmonkeys/rocker-ci:${TAG} --rm rocker-ci
