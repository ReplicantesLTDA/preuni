FROM alpine:3.19

RUN apk add --no-cache \
    postgresql16-client \
    s3cmd \
    bash \
    dcron \
    tzdata

COPY pg_dump_cron.sh /usr/local/bin/pg_dump_cron.sh
RUN chmod +x /usr/local/bin/pg_dump_cron.sh

# Schedule at 03:00 UTC daily.
RUN echo "0 3 * * * /usr/local/bin/pg_dump_cron.sh >> /var/log/backup.log 2>&1" \
    | crontab -

CMD ["crond", "-f", "-d", "8"]
