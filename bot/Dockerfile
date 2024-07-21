FROM oven/bun:latest

WORKDIR /app

COPY package.json bun.lockb ./
COPY src/ ./src/

RUN bun install

EXPOSE 5005

CMD ["bun", "run", "dev"]