FROM ubuntu:20.04
ADD dist/controller dist/repeater /usr/local/bin/
RUN mv /usr/local/bin/controller /usr/local/bin/gangway && \
    chmod +x /usr/local/bin/gangway && \
    chmod +x /usr/local/bin/repeater
CMD [ "gangway" ]