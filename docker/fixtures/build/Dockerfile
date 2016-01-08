FROM cloudrocker-base:latest
COPY droplet.tgz /app/
RUN chown vcap:vcap /app && cd /app && su vcap -c "tar zxf droplet.tgz" && rm droplet.tgz
EXPOSE 8080
USER vcap
WORKDIR /app
ENV HOME /app
ENV PORT 8080
ENV TMPDIR /app/tmp
CMD ["/bin/bash", "/app/cloudrocker-start-1c4352a23e52040ddb1857d7675fe3cc.sh", "/app", "the", "start", "command", "\"quoted", "string", "with", "spaces\""]
