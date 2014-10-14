FROM cloudfocker-base:latest
RUN /usr/sbin/useradd -mU -u 10000 -s /bin/bash vcap
COPY droplet.tgz /app/
RUN chown vcap:vcap /app && cd /app && su vcap -c "tar zxf droplet.tgz" && rm droplet.tgz
EXPOSE 8080
USER vcap
WORKDIR /app
ENV HOME /app
ENV PORT 8080
ENV TMPDIR /app/tmp
CMD ["/bin/bash", "/app/cloudfocker-start.sh", "/app", "test", "test", "test"]
