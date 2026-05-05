FROM debian:bookworm-slim

WORKDIR /app

COPY main ./
COPY configs/default.yml ./default.yml

RUN chmod +x ./main

EXPOSE 8082

CMD ["./main"]
