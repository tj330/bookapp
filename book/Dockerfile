FROM alpine:latest

WORKDIR /app

COPY main ./

COPY configs/default.yml ./

RUN chmod +x ./main

EXPOSE 8083

CMD ["/app/main"]
