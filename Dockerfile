FROM debian:stable
COPY bonos /usr/local/bin/bonos

# Create /data directory
WORKDIR /data
# Expose data volume
VOLUME /data

# Expose ports
EXPOSE 65139/tcp

# Set the default command
ENTRYPOINT [ "/usr/local/bin/bonos" ]
CMD [ "--db", "/data/bonos.db" ]
