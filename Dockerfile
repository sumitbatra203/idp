FROM scratch

COPY ./build_out/idp /

ENTRYPOINT ["/idp"]