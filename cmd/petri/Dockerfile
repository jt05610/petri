# base node image
FROM node:18-bullseye-slim as base

# set for base and all layer that inherit from it
ENV NODE_ENV production

# Install openssl for Prisma
# Install postgres for PostgreSQL
RUN apt-get update && apt-get install -y openssl postgresql

# Install all node_modules, including dev dependencies
FROM base as deps

WORKDIR /myapp

ADD package.json pnpm-lock.yaml .npmrc ./
RUN npm install --include=dev

# Setup production node_modules
FROM base as production-deps

WORKDIR /myapp

COPY --from=deps /myapp/node_modules /myapp/node_modules
ADD package.json pnpm-lock.yaml .npmrc ./
RUN npm prune --omit=dev

# Build the app
FROM base as build

WORKDIR /myapp

COPY --from=deps /myapp/node_modules /myapp/node_modules

ADD prisma .
RUN npx prisma generate
ADD . .
RUN npm run build

# Finally, build the production image with minimal footprint
FROM base

# PostgreSQL connection string (replace with your actual connection details)
ENV DATABASE_URL="postgresql://user:password@localhost:5432/petri"
ENV RABBITMQ_EXCHANGE="topic_devices"
ENV RABBITMQ_URI="amqp://guest:guest@localhost:5672/"
ENV SESSION_SECRET="super-duper-s3cret"
ENV PORT=3000
ENV NODE_ENV="production"

# Add shortcut for connecting to the PostgreSQL database CLI
RUN echo "#!/bin/sh\nset -x psql -U \$USER -h \$HOST -d \$DBNAME" > /usr/local/bin/database-cli && chmod +x /usr/local/bin/database-cli

WORKDIR /myapp

COPY --from=production-deps /myapp/node_modules /myapp/node_modules
COPY --from=build /myapp/node_modules/.prisma /myapp/node_modules/.prisma

COPY --from=build /myapp/build /myapp/build
COPY --from=build /myapp/public /myapp/public
COPY --from=build /myapp/package.json /myapp/package.json
COPY --from=build /myapp/start.sh /myapp/start.sh
COPY --from=build /myapp/prisma /myapp/prisma

ENTRYPOINT [ "./start.sh" ]