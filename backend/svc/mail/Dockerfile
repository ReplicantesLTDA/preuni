FROM elixir:1.17-otp-27-alpine

WORKDIR /app

# Install hex + rebar
RUN mix local.hex --force && mix local.rebar --force

# Install deps first (cache-friendly layer)
COPY mix.exs mix.lock ./
RUN MIX_ENV=prod mix deps.get --only prod

# Compile deps
RUN MIX_ENV=prod mix deps.compile

# Copy application source
COPY lib/ ./lib/
COPY config/ ./config/
COPY priv/ ./priv/

# Compile application
RUN MIX_ENV=prod mix compile

EXPOSE 4000

ENV PHX_SERVER=true
ENV MIX_ENV=prod

CMD ["mix", "phx.server"]
