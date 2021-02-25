FROM alpine
ADD dist/controller dist/repeater /usr/bin/
RUN mv /usr/bin/controller /usr/bin/gangway && \
    chmod +x /usr/bin/gangway && \
    chmod +x /usr/bin/repeater
CMD [ "gangway" ]